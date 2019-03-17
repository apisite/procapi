// b2pgx is a bridge from pgcall to pgx
package pgxpgcall

import (
	"github.com/apisite/pgcall/pgiface"
	"github.com/jackc/pgx"
)

// Config defines local application flags
type Config struct {
	Schema   string `long:"schema" env:"SCHEMA" default:"" description:"Database functions schema name or comma delimited list"`
	LogLevel string `long:"loglevel" env:"LOGLEVEL" default:"error" description:"DB logging level (trace|debug|info|warn|error|none)"`
	Retry    int    `long:"retry" default:"5" description:"Retry db connect after this interfal (secs), No retry if 0"`
	Workers  int    `long:"workers" default:"2" description:"DB connections count"`
}

type Rows struct {
	*pgx.Rows
}

func (r Rows) Columns() ([]string, error) {
	fields := r.FieldDescriptions()
	result := make([]string, len(fields))
	for k, v := range fields {
		result[k] = v.Name
	}
	return result, nil
}

type Result struct {
	pgx.CommandTag
}

func (r Result) RowsAffected() (int64, error) {
	return r.CommandTag.RowsAffected(), nil
}

type DB struct {
	*pgx.ConnPool
}

func (db DB) Exec(sql string, arguments ...interface{}) (pgiface.Result, error) {
	t, err := db.ConnPool.Exec(sql, arguments...)
	res := Result{CommandTag: t}
	return res, err
}

func (db DB) Query(sql string, args ...interface{}) (pgiface.Rows, error) {
	r, err := db.ConnPool.Query(sql, args...)
	rows := Rows{Rows: r}
	return rows, err
}
