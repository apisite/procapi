// +build !db

package pgxpgcall

import (
	"github.com/jackc/pgx"

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
