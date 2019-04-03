// +build db

package pgxpgcall

import (
	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/birkirb/loggers.v1"
)

type ServerSuite struct {
	suite.Suite
	cfg  Config
	srv  *DB
	hook *test.Hook
	log  loggers.Contextual
}

func (ss *ServerSuite) SetupSuite() {
	l, hook := test.NewNullLogger()
	ss.hook = hook
	l.SetLevel(logrus.DebugLevel)

	ss.log = mapper.NewLogger(l)
	hook.Reset()

	// Fill config with default values
	p := flags.NewParser(&ss.cfg, flags.Default|flags.IgnoreUnknown)
	//	_, err := p.ParseArgs([]string{})
	_, err := p.Parse() //Args([]string{})
	require.NoError(ss.T(), err)
	ss.srv, err = New(ss.cfg, ss.log)
	require.NoError(ss.T(), err)

}

func (ss *ServerSuite) TestQuery() {

	q := "select x from generate_series(0,2) x"
	rv, err := ss.srv.Query(q)
	require.NoError(ss.T(), err)
	want := []interface{}{int32(0), int32(1), int32(2)}
	assert.Equal(ss.T(), want, rv)
}

func (ss *ServerSuite) TestQueryProc() {

	q := "func_args"
	rv, err := ss.srv.QueryProc(q, "index")
	require.NoError(ss.T(), err)
	want := []map[string]interface{}{
		{"anno": "Схема БД", "arg": "a_nsp", "required": false, "type": "text"},
	}
	assert.Equal(ss.T(), want, rv)

}

/*
func TestRowsAffected(t *testing.T) {

	result := Result{CommandTag: "rows: 72"}
	var rows int64 = 72
	rv, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, rows, rv)
}
*/
/*

func expectTable(rows *MockRows, fields []string, maps []map[string]interface{}) {

	rows.EXPECT().Columns().
		Return(fields, nil)
	for _, m := range maps {
		var row []interface{}
		for _, f := range fields {
			row = append(row, m[f])
		}
		rows.EXPECT().Next().
			Return(true)
		rows.EXPECT().Values().
			Return(row, nil)
	}
	rows.EXPECT().Next().
		Return(false)
	rows.EXPECT().Err().
		Return(nil)
	rows.EXPECT().Close()

}


*/
