// +build db

package pgcall

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/apisite/pgcall/pgx-pgcall"
)

// Config holds all config vars
type TestConfig struct {
	PGFC Config      `group:"PGFC Options" namespace:"pgcall"`
	PGXI pgxpgcall.Config `group:"PG Options" namespace:"db" env-namespace:"DB"`
}

type ServerSuite struct {
	suite.Suite
	cfg  TestConfig
	srv  *Server
	hook *test.Hook
	req  *http.Request
}

func (ss *ServerSuite) SetupSuite() {

	// Fill config with default values
	p := flags.NewParser(&ss.cfg, flags.Default|flags.IgnoreUnknown)
	_, err := p.ParseArgs([]string{})
	//_, err := p.Parse() //Args([]string{})
	require.NoError(ss.T(), err)

	ss.cfg.PGXI.Schema = os.Getenv("DB_SCHEMA") // TODO: Parser have to get it
	l, hook := test.NewNullLogger()
	ss.hook = hook
	l.SetLevel(logrus.DebugLevel)
	log := mapper.NewLogger(l)

	hook.Reset()
	ss.cfg.PGXI.LogLevel = "debug"
	db, err := pgxpgcall.New(ss.cfg.PGXI, log) //, ss.cfg.LogLevel, ss.cfg.Schema, ss.cfg.Workers, ss.cfg.Retry)
	require.NoError(ss.T(), err)
	assert.Equal(ss.T(), "Added DB connection", ss.hook.LastEntry().Message)

	ss.req = httptest.NewRequest("GET", "http://example.com/foo", nil)
	//    w := httptest.NewRecorder()

	s, err := New(ss.cfg.PGFC, log, db)
	require.NoError(ss.T(), err)
	ss.srv = s
}

func (ss *ServerSuite) TestCall() {

	ss.hook.Reset()

	var allResult map[string][]map[string]interface{}
	helperLoadJSON(ss.T(), "result.json", &allResult)

	tests := []struct {
		name   string
		method string
		args   map[string]interface{}
		res    []map[string]interface{}
		err    string
	}{
		{name: "Res", method: "func_result", args: map[string]interface{}{"code": "index"}, res: allResult["index"]},
	}

	for _, tt := range tests {
		rv, err := ss.srv.Call(ss.req, tt.method, tt.args)
		require.NoError(ss.T(), err)
		assert.Equal(ss.T(), tt.res, rv)
	}
	/*
		for _, e := range ss.hook.Entries {
			fmt.Printf("ENT[%s]: %s\n", e.Level, e.Message)
		}
	*/
	// 3 debug lines about required arg (IN args, Use, Sql, Query)
	assert.Equal(ss.T(), 4, len(ss.hook.Entries))
	//assert.Equal(ss.T(), logrus.DebugLevel, ss.hook.LastEntry().Message)
}

func TestDBSuite(t *testing.T) {

	myTest := &ServerSuite{}
	suite.Run(t, myTest)
	/*
		myTest.hook.Reset()

		for _, e := range myTest.hook.Entries {
			fmt.Printf("ENT[%s]: %s\n", e.Level, e.Message)
		}
	*/
}
