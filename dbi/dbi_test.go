package dbi

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	_, err := Connect("", "")
	tn := "Connect without driver must return error"
	assert.NotNil(t, err, tn)
	assert.Equal(t, "sql: unknown driver \"\" (forgotten import?)", err.Error(), tn)
}

func TestBegin(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	assert.NotNil(t, db)
	defer db.Close()
	mdb := myDB{&sqlx.DB{DB: db}}

	e := fmt.Errorf("some error")

	mock.ExpectBegin().WillReturnError(e)
	tx, err := mdb.Beginx()
	assert.NotNil(t, err)
	assert.Equal(t, e, err)

	mock.ExpectBegin()
	tx, err = mdb.Beginx()
	assert.Nil(t, err)
	assert.NotNil(t, tx)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("select 1").WillReturnRows(rows)
	_, err = tx.Queryx("select 1")
	assert.Nil(t, err)

	result := sqlmock.NewErrorResult(e)
	mock.ExpectExec("^select 1").WillReturnResult(result)
	_, err = tx.Exec("select 1")
	assert.Nil(t, err)

}
