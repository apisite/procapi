// b2pgx is a bridge from pgcall to pgx
package pgxpgcall

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx"
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
	return &DB{dbh: dbh}, nil
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
}

func columns(r *pgx.Rows) ([]string, error) {
	fields := r.FieldDescriptions()
	result := make([]string, len(fields))
	for k, v := range fields {
		result[k] = v.Name
	}
	return result, nil
}

// QueryMaps fetches []map[string]interface{} from query result
func (db *DB) QueryMaps(query string, args ...interface{}) ([]map[string]interface{}, error) {
	r, err := db.dbh.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	result := []map[string]interface{}{}
	fields, _ := columns(r)
	for r.Next() {
		row, err := r.Values()
		if err != nil {
			return nil, err
		}
		rowMap := map[string]interface{}{}
		for k, v := range row {
			rowMap[fields[k]] = v
		}
		result = append(result, rowMap)
	}
	if r.Err() != nil {
		return nil, r.Err()
	}
	return result, nil
}

// Query fetches []interface{} from query result
func (db *DB) Query(query string, args ...interface{}) ([]interface{}, error) {
	r, err := db.dbh.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	result := []interface{}{}
	fields, _ := columns(r)
	if len(fields) != 1 {
		return nil, errors.New("single column must be returned")
	}
	for r.Next() {
		row, err := r.Values()
		if err != nil {
			return nil, err
		}
		result = append(result, row[0])
	}
	if r.Err() != nil {
		return nil, r.Err()
	}
	return result, nil
}
