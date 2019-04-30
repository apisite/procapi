// Package dbi implements DB interfaces based on sqlx.
package dbi

import (
	"github.com/jmoiron/sqlx"
)

// Result is a local copy of sql.Result for mocking
type Result interface {
	RowsAffected() (int64, error)
}

// Rows is a local copy of *sql[.Result for mocking
// Rows holds all of Rows methods used
type Rows interface {
	StructScan(dest interface{}) error
	MapScan(dest map[string]interface{}) error
	Scan(dest ...interface{}) error
	Close() error
	Next() bool
}

// Tx is a local copy of *sqlx.Tx for mocking
// Tx holds all of database transaction methods used
type Tx interface {
	Queryx(query string, args ...interface{}) (Rows, error)
	Exec(query string, args ...interface{}) (Result, error)
	Rollback() error
	Commit() error
}

// DB holds all of database methods, used for database mocking
type DB interface {
	Beginx() (Tx, error)
	Close() error
}

// myDB is a mockable *sqlx.DB
type myDB struct {
	*sqlx.DB
}

type myTx struct {
	*sqlx.Tx
}

// Connect return DB connection handle
func Connect(driverName, dataSourceName string) (DB, error) {
	dbh, err := sqlx.Connect(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return myDB{dbh}, nil
}

// Beginx begins a transaction and returns an Tx instead of an *sqlx.Tx.
func (db myDB) Beginx() (Tx, error) {
	tx, err := db.DB.Beginx()
	if err != nil {
		return nil, err
	}
	return myTx{tx}, nil
}

// Queryx within a transaction, returns Rows instead *sqlx.Rows
func (tx myTx) Queryx(query string, args ...interface{}) (Rows, error) {
	return tx.Tx.Queryx(query, args...)
}

// Exec within a transaction, returns Result instead sql.Rows
func (tx myTx) Exec(query string, args ...interface{}) (Result, error) {
	return tx.Tx.Exec(query, args...)
}
