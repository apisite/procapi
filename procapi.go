package procapi

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"sync"

	"github.com/jackc/pgx/v4"
	"gopkg.in/birkirb/loggers.v1"
)

// codebeat:disable[TOO_MANY_IVARS]

// Config defines local application flags
type Config struct {
	DSN           string    `long:"dsn" default:"postgres://?sslmode=disable" description:"Database connect string"`
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

// Service holds API service methods
type Service struct {
	dbh          *pgx.Conn
	config       Config
	log          loggers.Contextual
	methods      map[string]Method
	mx           sync.RWMutex
	typeM        Marshaller
	schemaSuffix string
}

// codebeat:enable[TOO_MANY_IVARS]

// Marshaller holds methods for database values marshalling
type Marshaller interface {
	Marshal(typ string, v interface{}) (interface{}, error)
	Unmarshal(typ string, data interface{}) (rv interface{}, err error)
}


//Functional options
//https://github.com/tmrts/go-patterns/blob/master/idiom/functional-options.md

// Option is a functional options return type
type Option func(*Service)

// Marshall allows to change default marshaller
func Marshall(m Marshaller) Option {
	return func(srv *Service) {
		srv.typeM = m
	}
}

// New returns procapi service
func New(cfg Config, log loggers.Contextual, dbh *pgx.Conn, options ...Option) *Service {
	srv := &Service{
		log: log, config: cfg, dbh: dbh,
	}
	for _, option := range options {
		option(srv)
	}

	if srv.typeM == nil {
		srv.typeM = &PGType{}
	}
	return srv
}

// Method returns method by name
func (srv *Service) Method(name string) (Method, bool) {
	srv.mx.RLock()
	defer srv.mx.RUnlock()
	m, ok := srv.methods[name]
	return m, ok
}

// LoadMethodsTx load methods within given transaction for nsp if given, all of methods otherwise
func (srv *Service) LoadMethodsTx(tx pgx.Tx) error {
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
func (srv *Service) FetchMethods(tx pgx.Tx, nsp *[]string) (map[string]Method, error) {
	const SQL = "select * from %s.%s(%s)"
	schema := srv.config.FuncSchema
	if srv.schemaSuffix != "" {
		schema += "_" + srv.schemaSuffix
	}
	ctx := context.Background()
	rows, err := tx.Query(ctx, fmt.Sprintf(SQL, schema, srv.config.IndexFunc, positionalArgs(nsp)), nsp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rvTemp := []Method{}
	for rows.Next() {
		r := Method{}
		err := ScanStruct(rows, &r)
		if err != nil {
			return nil, err
		}
		rvTemp = append(rvTemp, r)
	}
	rv := map[string]Method{}
	for _, v := range rvTemp {
		k := v.Name
		rows, err := tx.Query(ctx, fmt.Sprintf(SQL, schema, srv.config.InDefFunc, positionalArgs(k)), k)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		inArgs := map[string]InDef{}
		for rows.Next() {
			r := InDef{}
			err := ScanStruct(rows, &r)
			if err != nil {
				return nil, err
			}
			inArgs[strings.TrimPrefix(r.Name, srv.config.ArgTrimPrefix)] = r
		}
		v.In = inArgs

		rows, err = tx.Query(ctx, fmt.Sprintf(SQL, schema, srv.config.OutDefFunc, positionalArgs(k)), k)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			r := OutDef{}
			err := ScanStruct(rows, &r)
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
	ctx := context.Background()
	tx, err := dbh.Begin(ctx)
	if err != nil {
		return nil, err
	}
	var rv interface{}
	rv, err = srv.CallTx(tx, method, args)
	if err != nil { // TODO: or Method.IsRo
		tx.Rollback(ctx)
	} else {
		err = tx.Commit(ctx)
	}
	return rv, err
}

// CallTx calls postgresql stored function within given transaction
func (srv *Service) CallTx( //r *http.Request,
	tx pgx.Tx,
	method string,
	args map[string]interface{},
) (interface{}, error) {
	// Check for Marshaller is set
	if srv.typeM == nil {
		return nil, &callError{code: errNilMarshaller}
	}
	// Lookup method.
	methodSpec, ok := srv.Method(method)
	if !ok {
		return nil, (&callError{code: errNotFound}).addContext("name", method)
	}

	var missedArgs []string
	var inAssigns []string
	var inVars []interface{}
	var err error

	if methodSpec.In != nil {
		missedArgs, inAssigns, inVars, err = srv.namedArgs(methodSpec.In, args)
	}
	if err != nil {
		return nil, err
	}
	if len(missedArgs) > 0 {
		return nil, (&callError{code: errArgsMissed}).addContext("args", missedArgs)
	}

	ctx := context.Background()
	if methodSpec.Result != nil && *methodSpec.Result == "void" {
		// no data returned
		sql := fmt.Sprintf("SELECT %s.%s(%s)",
			methodSpec.Class,
			methodSpec.Func,
			strings.Join(inAssigns, ", "),
		)
		qr, err := tx.Exec(ctx, sql, inVars...)
		ctra := qr.RowsAffected()
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
	// fmt.Printf(">>> row------------------: %+v %v\n", sql, inVars)

	rows, err := tx.Query(ctx, sql, inVars...)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	defer rows.Close()

	return fetch(methodSpec, srv.typeM, rows)
}

func fetch(methodSpec Method, mars Marshaller, rows pgx.Rows) (interface{}, error) {

	var rv []interface{}
	var err error

	for rows.Next() {
		var r interface{}
		if methodSpec.IsStruct {
			m := map[string]interface{}{}
			err = ScanMap(rows, &m)

			if err != nil {
				return nil, err
			}
			for _, c := range methodSpec.Out {
				if m[c.Name] == nil {
					continue
				}
				out, err := mars.Unmarshal(c.Type, m[c.Name])
				if err != nil {
					return nil, err
				}
				m[c.Name] = out
			}
			r = m
		} else {
			// get 1st column only
			var rr []interface{}
			rr, err = rows.Values()
			if err != nil {
				return nil, err
			}
			r = rr[0]
		}
		rv = append(rv, r)
	}

	if !methodSpec.IsSet {
		if len(rv) != 1 {
			return nil, &callError{code: errNotSingleRV}
		}
		rv0 := rv[0]
		return &rv0, nil // might be null
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
	err error,
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
		out, e := srv.typeM.Marshal(v.Type, a)
		if e != nil {
			err = (&callError{code: errArgCast}).
				addContext("arg", v.Name).
				addContext("type", v.Type).
				addContext("val", a).
				addContext("err", e)
			return
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
