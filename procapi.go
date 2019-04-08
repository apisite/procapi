package procapi

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/apisite/procapi/pgtype"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gopkg.in/birkirb/loggers.v1"
)

// Config defines local application flags
type Config struct {
	DSN           string    `long:"dsn" default:"" description:"Database connect string, i.e. postgres://user:pass@host/dbname?sslmode=disable"`
	Driver        string    `long:"driver" default:"postgres" description:"Database driver"`
	InDefFunc     string    `long:"indef" default:"func_args" description:"Argument definition function"`
	OutDefFunc    string    `long:"outdef" default:"func_result" description:"Result row definition function"`
	IndexFunc     string    `long:"index" default:"index" description:"Available functions list"`
	FuncSchema    string    `long:"schema" default:"rpc" description:"Definition functions schema"`
	ArgSyntax     string    `long:"arg_syntax" default:":=" description:"Default named args syntax (:= or =>)"`
	ArgTrimPrefix string    `long:"arg_prefix" default:"a_" description:"Trim prefix from arg name"`
	NameSpaces    *[]string `long:"nsp" description:"Proc namespace(s)"`

	// TODO: Lim*
	// LimArg  - если у ф-и нет этого аргумента, добавить в запрос `LIMIT LimDefault`
	// LimDefault - лимит строк по умолчанию

}

// InDef holds function argument attributes
type InDef struct {
	Name     string  `db:"arg"`
	Type     string  `db:"type"`
	Required bool    `db:"required"` // TODO: is_required
	Default  *string `db:"def_val" json:",omitempty"`
	Anno     *string `db:"anno" json:",omitempty"`
	// Check    string `json:"check,omitempty" sql:"-"` // validate argument
}

// OutDef holds function result attributes
type OutDef struct {
	Name string  `db:"arg"`
	Type string  `db:"type"`
	Anno *string `db:"anno" json:",omitempty"`
}

// Method holds method attributes
type Method struct {
	Name     string           `db:"code"`
	Class    string           `db:"nspname"`
	Func     string           `db:"proname"`
	Anno     string           `db:"anno"`
	IsRO     bool             `db:"is_ro"`
	IsSet    bool             `db:"is_set"`
	IsStruct bool             `db:"is_struct"`
	Sample   *string          `db:"sample" json:",omitempty"`
	Result   *string          `db:"result" json:",omitempty"`
	In       map[string]InDef //`json:",omitempty"`
	Out      []OutDef         //`json:",omitempty"`
}

type Result interface {
	RowsAffected() (int64, error)
}

type Rows interface {
	StructScan(dest interface{}) error
	MapScan(dest map[string]interface{}) error
	Scan(dest ...interface{}) error
	Close() error
	Next() bool
}

// Tx holds all of database transaction methods used
type Tx interface {
	Queryx(query string, args ...interface{}) (Rows, error)
	Exec(query string, args ...interface{}) (Result, error)
	Rollback() error
	Commit() error
}

// DB holds all of database methods used
type DB interface {
	Beginx() (Tx, error)
	Close() error
}

// DB holds all of database methods used
type MyDB struct {
	*sqlx.DB
}

func (db MyDB) Beginx() (Tx, error) {
	tx, err := db.DB.Beginx()
	if err != nil {
		return nil, err
	}
	return myTx{tx}, nil
}
func (db MyDB) Close() error {
	return db.DB.Close()
}

type myTx struct {
	*sqlx.Tx
}

func (tx myTx) Queryx(query string, args ...interface{}) (Rows, error) {
	rows, err := tx.Tx.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	return myRows{rows}, nil
}
func (tx myTx) Exec(query string, args ...interface{}) (Result, error) {
	res, err := tx.Tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return myResult{res}, nil
}

type myRows struct {
	*sqlx.Rows
}

type myResult struct {
	sql.Result
}

/*
func (res MResult) RowsAffected() (int64, error) {
	return res.Result.RowsAffected()
}
*/

// Marshalle rholds all of database methods used
type Marshaller interface {
	Marshal(typ string, v interface{}) (interface{}, error)
	Unmarshal(typ string, data interface{}) (rv interface{}, err error)
}

// Service holds API service methods
type Service struct {
	dbh          DB
	config       Config
	log          loggers.Contextual
	methods      map[string]Method
	mx           sync.RWMutex
	typeM        Marshaller
	schemaSuffix string
}

// New returns procapi service
func New(cfg Config, log loggers.Contextual, dbh DB) *Service {
	srv := Service{log: log, config: cfg, dbh: dbh, typeM: pgtype.New()}
	return &srv
}

// SetSchemaSuffix sets suffix for all of used shemas names
func (srv *Service) SetSchemaSuffix(suffix string) *Service {
	srv.mx.Lock()
	defer srv.mx.Unlock()
	srv.schemaSuffix = suffix
	return srv
}

// SetMarshaller allows to change internal pgtype Marshaller to another one
func (srv *Service) SetMarshaller(m Marshaller) *Service {
	srv.mx.Lock()
	defer srv.mx.Unlock()
	srv.typeM = m
	return srv
}

// Open opens DB connection
func (srv *Service) Open() error {
	srv.mx.RLock()
	dbh := srv.dbh
	srv.mx.RUnlock()
	if dbh != nil {
		return errors.New("dbh opened alreade")
	}
	dsn := srv.config.DSN
	if dsn == "" {
		// Use postgresql ENV vars
		dsn = "postgres://?sslmode=disable"
	}
	conn, err := sqlx.Connect(srv.config.Driver, dsn)
	if err != nil {
		return err
	}
	srv.mx.Lock()
	defer srv.mx.Unlock()
	srv.dbh = MyDB{conn}
	return nil
}

// Method returns method by name
func (srv *Service) DB() DB {
	srv.mx.RLock()
	defer srv.mx.RUnlock()
	return srv.dbh
}

// Method returns method by name
func (srv *Service) Method(name string) (Method, bool) {
	srv.mx.RLock()
	defer srv.mx.RUnlock()
	m, ok := srv.methods[name]
	return m, ok
}

// LoadMethods load methods for nsp if given, all of methods otherwise
func (srv *Service) LoadMethods() error {
	srv.mx.RLock()
	dbh := srv.dbh
	srv.mx.RUnlock()
	if dbh == nil {
		return errors.New("dbh must be not nil")
	}
	tx, err := dbh.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() // No DB changes here
	return srv.LoadMethodsTx(tx)
}

// LoadMethodsTx load methods within given transaction for nsp if given, all of methods otherwise
func (srv *Service) LoadMethodsTx(tx Tx) error {
	rv, err := srv.FetchMethods(tx, srv.config.NameSpaces)
	if err != nil {
		return err
	}
	srv.mx.Lock()
	srv.methods = rv
	srv.mx.Unlock()
	return nil
}

// FetchMethods fetches from DB methods definition for given namespaces
func (srv *Service) FetchMethods(tx Tx, nsp *[]string) (map[string]Method, error) {
	const SQL = "select * from %s.%s(%s)"
	schema := srv.config.FuncSchema
	if srv.schemaSuffix != "" {
		schema += "_" + srv.schemaSuffix
	}
	rows, err := tx.Queryx(fmt.Sprintf(SQL, schema, srv.config.IndexFunc, positionalArgs(nsp)), nsp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rvTemp := []Method{}
	for rows.Next() {
		r := Method{}
		err := rows.StructScan(&r)
		if err != nil {
			return nil, err
		}
		rvTemp = append(rvTemp, r)
	}
	rv := map[string]Method{}
	for _, v := range rvTemp {
		k := v.Name
		rows, err := tx.Queryx(fmt.Sprintf(SQL, schema, srv.config.InDefFunc, positionalArgs(k)), k)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		inArgs := map[string]InDef{}
		for rows.Next() {
			r := InDef{}
			err := rows.StructScan(&r)
			if err != nil {
				return nil, err
			}
			inArgs[strings.TrimPrefix(r.Name, srv.config.ArgTrimPrefix)] = r
		}
		v.In = inArgs

		rows, err = tx.Queryx(fmt.Sprintf(SQL, schema, srv.config.OutDefFunc, positionalArgs(k)), k)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			r := OutDef{}
			err := rows.StructScan(&r)
			if err != nil {
				return nil, err
			}
			v.Out = append(v.Out, r)
		}
		rv[k] = v
	}
	return rv, nil
}

// Call calls postgresql stored function
func (srv *Service) Call(
	r *http.Request, // TODO: interface to session & request data
	method string,
	args map[string]interface{},
) (interface{}, error) {
	srv.mx.RLock()
	dbh := srv.dbh
	srv.mx.RUnlock()
	if dbh == nil {
		return nil, errors.New("dbh must be not nil")
	}
	tx, err := dbh.Beginx()
	if err != nil {
		return nil, err
	}
	var rv interface{}
	rv, err = srv.CallTx(tx, method, args)
	if err != nil { // TODO: or Method.IsRo
		tx.Rollback()
	} else {
		err = tx.Commit()
	}
	return rv, err
}

// CallTx calls postgresql stored function within given transaction
func (srv *Service) CallTx( //r *http.Request,
	tx Tx,
	method string,
	args map[string]interface{},
) (interface{}, error) {
	// Lookup method.
	methodSpec, ok := srv.Method(method)
	if !ok {
		return nil, (&callError{code: errNotFound}).addContext("name", method)
	}

	var missedArgs []string
	var inAssigns []string
	var inVars []interface{}

	if methodSpec.In != nil {
		missedArgs, inAssigns, inVars = srv.namedArgs(methodSpec.In, args)
	}
	if len(missedArgs) > 0 {
		return nil, (&callError{code: errArgsMissed}).addContext("args", missedArgs)
	}

	if methodSpec.Out == nil && methodSpec.Result == nil {
		// no data returned
		sql := fmt.Sprintf("SELECT %s.%s(%s)",
			methodSpec.Class,
			methodSpec.Func,
			strings.Join(inAssigns, ", "),
		)
		qr, err := tx.Exec(sql, inVars...)
		ctra, _ := qr.RowsAffected()
		srv.log.Debugf("Rows affected: %d", ctra) // TODO: Header.Add ?
		return nil, err
	}

	var outCols []string
	if methodSpec.Out != nil {
		for _, v := range methodSpec.Out {
			outCols = append(outCols, v.Name)
		}
	}

	from := ""
	if len(outCols) > 0 {
		from = " from "
	}

	sql := fmt.Sprintf("select %s%s%s.%s(%s)",
		strings.Join(outCols, ", "),
		from,
		methodSpec.Class,
		methodSpec.Func,
		strings.Join(inAssigns, ", "),
	)
	srv.log.Debugf("sql: %s, args: %v\n", sql, inVars)
	var rv []interface{}
	var err error

	rows, err := tx.Queryx(sql, inVars...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var r interface{}
		if methodSpec.IsStruct {
			m := map[string]interface{}{}
			err = rows.MapScan(m)
			if err != nil {
				return nil, err
			}
			for _, c := range methodSpec.Out {
				if m[c.Name] == nil {
					continue
				}
				out, err := srv.typeM.Unmarshal(c.Type, m[c.Name])
				if err != nil {
					return nil, err
				}
				m[c.Name] = out
			}
			r = m
		} else {
			// get 1st column only
			err = rows.Scan(&r)
		}
		if err != nil {
			return nil, err
		}
		rv = append(rv, r)
	}

	if !methodSpec.IsSet {
		if len(rv) != 1 {
			return nil, errors.New("single row must be returned")
		}
		return rv[0], nil
	}
	return rv, nil
}

// namedArgs returns data for building proc call with named args
func (srv *Service) namedArgs(
	inDef map[string]InDef,
	args map[string]interface{},
) (
	missedArgs []string,
	inAssigns []string,
	inVars []interface{},
) {
	log := srv.log
	log.Debugf("IN args: %+v", inDef)
	for k, v := range inDef {
		a, ok := args[k]
		if !ok {
			if v.Required {
				missedArgs = append(missedArgs, k)
			} else {
				log.Debugf("Skip missed value of %s", k)
			}
			continue
		}
		if reflect.ValueOf(a).Kind() == reflect.Ptr {
			if reflect.ValueOf(a).IsNil() {
				if v.Required {
					missedArgs = append(missedArgs, k)
					continue
				} else {
					log.Debugf("Use NULL for empty ref of %s", k)
				}
			} else {
				a = reflect.ValueOf(a).Elem().Interface() // dereference ptr
			}
		}
		inAssigns = append(inAssigns, fmt.Sprintf("%s %s $%d", v.Name, srv.config.ArgSyntax, len(inAssigns)+1))
		out, err := srv.typeM.Marshal(v.Type, a)
		if err != nil {
			log.Debugf("Use: %s (%+v)", k, a)
			continue
		}
		a = out

		inVars = append(inVars, a)
		log.Debugf("Use: %s (%+v)", k, a)
	}
	return
}

// positionalArgs returns string with function args placeholders
func positionalArgs(args ...interface{}) string {
	inAssigns := make([]string, len(args))
	for i := range args {
		inAssigns[i] = fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(inAssigns, ", ")
}
