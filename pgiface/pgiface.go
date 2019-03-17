package pgiface

// sql.Rows does not implement pgiface.Rows (wrong type for Close method)
//*sql.Rows does not implement pgiface.Rows (missing Values method)

type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...interface{}) (err error)
	Values() ([]interface{}, error)

	Columns() ([]string, error)
}

// Result interface holds a common part for
// * https://golang.org/pkg/database/sql/#Result
// * https://godoc.org/github.com/jackc/pgx#CommandTag
type Result interface {
	RowsAffected() (int64, error)
}

type DB interface {
	Exec(sql string, arguments ...interface{}) (result Result, err error)
	Query(sql string, args ...interface{}) (Rows, error)
}
