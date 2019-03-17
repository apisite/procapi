package pgxpgcall

import (
	"github.com/jackc/pgx"
	"os"

	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/birkirb/loggers.v1"
)

type ServerSuite struct {
	suite.Suite
	srv  *pgx.ConnPool
	hook *test.Hook
	log  loggers.Contextual
}

func (ss *ServerSuite) SetupSuite() {
	l, hook := test.NewNullLogger()
	ss.hook = hook
	l.SetLevel(logrus.DebugLevel)
	ss.log = mapper.NewLogger(l)
	hook.Reset()
}

func (ss *ServerSuite) TestLogLevelError() {
	cfg := Config{LogLevel: "no"}
	_, err := New(cfg, ss.log)
	require.NotNil(ss.T(), err)
	assert.Equal(ss.T(), "Unable to parse log level no: invalid log level", err.Error())
}

func (ss *ServerSuite) TestNewServerError() {
	ss.hook.Reset()
	tests := []struct {
		name  string
		env   string
		value string
		err   string
	}{
		{name: "BadPortSyntax", env: "PGPORT", value: "GoLangCode", err: "Unable to parse environment: strconv.ParseUint: parsing \"GoLangCode\": invalid syntax"},
		{name: "BadPortValue", env: "PGPORT", value: "1", err: "dial tcp 127.0.0.1:1: connect: connection refused"},
	}

	cfg := Config{LogLevel: "none", Workers: 1}
	for _, tt := range tests {
		pre := os.Getenv(tt.env)
		os.Setenv(tt.env, tt.value)
		defer func() {
			os.Setenv(tt.env, pre)
		}()
		_, err := New(cfg, ss.log)
		require.NotNil(ss.T(), err)
		assert.Equal(ss.T(), tt.err, err.Error())
	}

	assert.Equal(ss.T(), 1, len(ss.hook.Entries))
	assert.Equal(ss.T(), logrus.WarnLevel, ss.hook.LastEntry().Level)
	assert.Equal(ss.T(), "DB connect failed: "+tests[1].err, ss.hook.LastEntry().Message)

}

func (ss *ServerSuite) TestLog() {
	ss.hook.Reset()
	log := Logger{l: ss.log}
	tests := []struct {
		name  string
		level pgx.LogLevel
		data  map[string]interface{}
	}{
		{name: "LogLevelTrace", level: pgx.LogLevelTrace},
		{name: "LogLevelDebug", level: pgx.LogLevelDebug},
		{name: "LogLevelInfo", level: pgx.LogLevelInfo},
		{name: "LogLevelWarn", level: pgx.LogLevelWarn},
		{name: "LogLevelError", level: pgx.LogLevelError},
		{name: "LogLevelNone", level: pgx.LogLevelNone},
	}

	for _, tt := range tests {
		log.Log(tt.level, tt.name, map[string]interface{}{})
		assert.Equal(ss.T(), tt.name, ss.hook.LastEntry().Message)
	}

	assert.Equal(ss.T(), len(tests), len(ss.hook.Entries))
	assert.Equal(ss.T(), logrus.ErrorLevel, ss.hook.LastEntry().Level)
}
