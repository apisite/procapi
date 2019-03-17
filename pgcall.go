// Package pgcall holds pg functions call methods
package pgcall

import (
	//	"fmt"
	//	"net/http"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/birkirb/loggers.v1"

	"github.com/apisite/pgcall/pgiface"
)

// Config defines local application flags
type Config struct {
	InDefFunc     string `long:"indef" default:"func_args" description:"Argument definition function"`
	OutDefFunc    string `long:"outdef" default:"func_result" description:"Result row definition function"`
	IndexFunc     string `long:"index" default:"index" description:"Available functions list"`
	ArgSyntax     string `long:"arg_syntax" default:":=" description:"Default named args syntax (:= or =>)"`
	ArgTrimPrefix string `long:"arg_prefix" default:"a_" description:"Trim prefix from arg name"`

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
	Name     string            `db:"code"`
	Class    string            `db:"nspname"`
	Func     string            `db:"proname"`
	Anno     string            `db:"anno"`
	IsRO     bool              `db:"is_ro"`
	IsSet    bool              `db:"is_set"`
	IsStruct bool              `db:"is_struct"`
	Sample   *string           `db:"sample" json:",omitempty"`
	Result   *string           `db:"result" json:",omitempty"`
	In       *map[string]InDef `json:",omitempty"`
	Out      *[]OutDef         `json:",omitempty"`
}

// Server holds RPC methods
type Server struct {
	dbh     pgiface.DB //*pgx.ConnPool
	config  Config
	log     loggers.Contextual
	methods *map[string]Method
	mux     sync.RWMutex
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
func New(cfg Config, log loggers.Contextual, dbh pgiface.DB) (*Server, error) {
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

// Methods returns methods map
func (srv *Server) Methods() *map[string]Method {
	srv.mux.RLock()
	defer srv.mux.RUnlock()
	return srv.methods
}

// MethodIsRO returns true if method exists and read-only
func (srv *Server) MethodIsRO(method string) bool {
	srv.mux.RLock()
	methods := *srv.methods
	srv.mux.RUnlock()
	m, ok := methods[method]
	if !ok {
		return false
	}
	return m.IsRO
}

func (srv *Server) LoadMethods(nsp *string) error {

	cfg := srv.config

	m, err := srv.CallMapAny(cfg.IndexFunc)
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

		args, err := srv.CallMapAny(cfg.InDefFunc, result.Name)
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
		result.In = &inArgs

		outs, err := srv.CallMapAny(cfg.OutDefFunc, result.Name)
		if err != nil {
			return err
		}
		outArgs := []OutDef{}
		for _, vOut := range outs {
			var out OutDef
			err = Decode(vOut, &out)
			if err != nil {
				return err
			}
			outArgs = append(outArgs, out)
		}
		result.Out = &outArgs

		re[result.Name] = result
	}
	srv.mux.Lock()
	srv.methods = &re
	srv.mux.Unlock()
	return nil

	/*
		out, err := json.MarshalIndent(re, "", "  ")
		fmt.Printf(">> %s\n", out)

		var mg map[string]Method
		helperLoadJSON(ss.T(), "methods.golden.json", &mg)
		assert.Equal(ss.T(), mg, re)
	*/

}

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
