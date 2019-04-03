// Package pgcall implements a caller of postgresql stored functions, which intended to use via http and in templates.
package pgcall

import (
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/birkirb/loggers.v1"
)

// Config defines local application flags
type Config struct {
	InDefFunc     string `long:"indef" default:"func_args" description:"Argument definition function"`
	OutDefFunc    string `long:"outdef" default:"func_result" description:"Result row definition function"`
	IndexFunc     string `long:"index" default:"index" description:"Available functions list"`
	ArgSyntax     string `long:"arg_syntax" default:":=" description:"Default named args syntax (:= or =>)"`
	ArgTrimPrefix string `long:"arg_prefix" default:"a_" description:"Trim prefix from arg name"`

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

// DB holda all of database methods used (see pgxpgcall)
type DB interface {
	QueryProc(method string, args ...interface{}) ([]map[string]interface{}, error)
	Exec(sql string, arguments ...interface{}) (int64, error)
	QueryMaps(sql string, args ...interface{}) ([]map[string]interface{}, error)
	Query(sql string, args ...interface{}) ([]interface{}, error)
}

// Server holds RPC methods
type Server struct {
	dbh     DB
	config  Config
	log     loggers.Contextual
	methods map[string]Method
	mx      sync.RWMutex
}

/*
type PGCallerConfig interface {
	Log() loggers.Contextual
	Config() *Config
	DB() *db
	Cache()
	Session()
	Meta()
	NativeAPI() // map[pkg]Class - методы для !IsSQL
}
[pgcall.]New(cfg PGCallerConfig) *PGCaller
*/

// New returns pgcall server object
func New(cfg Config, log loggers.Contextual, dbh DB) (*Server, error) {
	if dbh == nil {
		return nil, errors.New("dbh must be not nil")
	}
	srv := Server{log: log, config: cfg, dbh: dbh}
	err := srv.LoadMethods(nil)
	if err != nil {
		return nil, err
	}
	return &srv, nil
}

// Method returns method by name
func (srv *Server) Method(name string) (Method, bool) {
	srv.mx.RLock()
	defer srv.mx.RUnlock()
	m, ok := srv.methods[name]
	return m, ok
}

// LoadMethods load methods for nsp if given, all of methods otherwise
func (srv *Server) LoadMethods(nsp *string) error {

	cfg := srv.config

	m, err := srv.dbh.QueryProc(cfg.IndexFunc)
	if err != nil {
		return err
	}

	re := map[string]Method{}

	for _, v := range m {

		var result Method
		err := Decode(v, &result)
		if err != nil {
			return err
		}

		args, err := srv.dbh.QueryProc(cfg.InDefFunc, result.Name)
		if err != nil {
			return err
		}
		inArgs := map[string]InDef{}
		for _, vIn := range args {
			var r InDef
			err = Decode(vIn, &r)
			if err != nil {
				return err
			}
			inArgs[strings.TrimPrefix(r.Name, cfg.ArgTrimPrefix)] = r
		}
		result.In = inArgs

		outs, err := srv.dbh.QueryProc(cfg.OutDefFunc, result.Name)
		if err != nil {
			return err
		}
		for _, vOut := range outs {
			var out OutDef
			err = Decode(vOut, &out)
			if err != nil {
				return err
			}
			result.Out = append(result.Out, out)
		}
		re[result.Name] = result
	}
	srv.mx.Lock()
	srv.methods = re
	srv.mx.Unlock()
	return nil

	/*
		out, err := json.MarshalIndent(re, "", "  ")
		fmt.Printf(">> %s\n", out)

		var mg map[string]Method
		helperLoadJSON(ss.T(), "methods.golden.json", &mg)
		assert.Equal(ss.T(), mg, re)
	*/

}

// Decode fills struct from map using ithub.com/mitchellh/mapstructure
func Decode(input interface{}, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   output,
		TagName:  "db",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
