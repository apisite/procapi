// Package pgxpgcall implements a jackc/pgx backend for pgcall.
// pgx was choosen for its postgresql type support
package pgxpgcall

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/pkg/errors"
	"gopkg.in/birkirb/loggers.v1"
)

// Config defines local application flags
type Config struct {
	Schema   string `long:"schema" env:"SCHEMA" default:"" description:"Database functions schema name or comma delimited list"`
	LogLevel string `long:"loglevel" env:"LOGLEVEL" default:"error" description:"DB logging level (trace|debug|info|warn|error|none)"`
	Retry    int    `long:"retry" default:"5" description:"Retry db connect after this interfal (secs), No retry if 0"`
	Workers  int    `long:"workers" default:"2" description:"DB connections count"`
}

type DB struct {
	dbh *pgx.ConnPool
	cfg Config
}

func New(cfg Config, log loggers.Contextual) (*DB, error) {
	config, err := initPool(cfg, log)
	if err != nil {
		return nil, err
	}
	var dbh *pgx.ConnPool
	for {
		dbh, err = pgx.NewConnPool(*config)
		if err == nil {
			break
		}
		log.Warnf("DB connect failed: %v", err)
		if cfg.Retry == 0 {
			break
		}
		time.Sleep(time.Second * time.Duration(cfg.Retry)) // sleep & repeat
	}
	if err != nil {
		return nil, err
	}
	return &DB{dbh: dbh, cfg: cfg}, nil
}

func (db *DB) Exec(sql string, arguments ...interface{}) (int64, error) {
	t, err := db.dbh.Exec(sql, arguments...)
	if err == nil {
		return t.RowsAffected(), nil
	}
	return 0, err
}

// QueryProc calls postgresql stored function without metadata usage
func (db *DB) QueryProc(method string, args ...interface{}) ([]map[string]interface{}, error) {
	inAssigns := make([]string, len(args))
	for i := range args {
		inAssigns[i] = fmt.Sprintf("$%d", i+1)
	}
	sql := fmt.Sprintf("select * from %s(%s)",
		method,
		strings.Join(inAssigns, ", "),
	)
	return db.QueryMaps(sql, args...)
	/*
		rv, err := db.QueryMaps(sql, args...)
		code := ""
		if len(args) > 0 {
			code = "." + args[0].(string)
		}
		checkTestDataUpdate(method+code, rv)
		return rv, err
	*/
}

// QueryMaps fetches []map[string]interface{} from query result
func (db *DB) QueryMaps(query string, args ...interface{}) ([]map[string]interface{}, error) {

	tx, err := db.dbh.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		tx.Rollback()
	}()

	r, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	result := []map[string]interface{}{}
	fields := r.FieldDescriptions()
	for r.Next() {
		row, err := r.Values()
		if err != nil {
			return nil, err
		}
		rowMap := map[string]interface{}{}
		for k, v := range row {
			if v == nil {
				continue
			}
			val, err := decodeValue(fields[k].DataTypeName, v)
			if err != nil {
				return nil, err
			}
			rowMap[fields[k].Name] = val
		}
		result = append(result, rowMap)
	}
	if r.Err() != nil {
		return nil, r.Err()
	}
	tx.Commit()
	return result, nil
}

// Query fetches []interface{} (slece of 1st column values) from query result
func (db *DB) Query(query string, args ...interface{}) ([]interface{}, error) {
	r, err := db.dbh.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	result := []interface{}{}
	fields := r.FieldDescriptions()
	if len(fields) != 1 {
		return nil, errors.Errorf("single column must be returned (got %d)", len(fields))
	}
	for r.Next() {
		row, err := r.Values()
		if err != nil {
			return nil, err
		}
		v := row[0]
		if v == nil {
			result = append(result, v)
			continue
		}
		val, err := decodeValue(fields[0].DataTypeName, v)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}
	if r.Err() != nil {
		return nil, r.Err()
	}
	return result, nil
}

// decode value from pgx.pgtype to go type
func decodeValue(typ string, val interface{}) (interface{}, error) {
	if strings.HasPrefix(typ, "_") {
		// value is slice
		switch typ {
		case "_int4":
			rv := []int32{}
			err := val.(*pgtype.Int4Array).AssignTo(&rv)
			return rv, err
		case "_text":
			rv := []string{}
			err := val.(*pgtype.TextArray).AssignTo(&rv)
			return rv, err
		default:
			return nil, errors.Errorf("result of type %s does not supported", typ)
		}
	}
	switch typ {
	case "numeric":
		var rv float32
		err := val.(pgtype.Value).AssignTo(&rv)
		return rv, err
	case "interval": // TODO: interval with months or days cannot be decoded into *time.Duration
		var rv time.Duration
		err := val.(pgtype.Value).AssignTo(&rv)
		return rv, err
	}
	return val, nil
}
