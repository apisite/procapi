// +build !nocommon1

package pgcall

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func (ss *ServerSuite) TestMethods() {
	m := ss.srv.Methods()
	var mg map[string]Method

	helperLoadJSON(ss.T(), "methods.golden.json", &mg)
	checkTestUpdate("methods.golden.json", m)
	assert.Equal(ss.T(), mg, m)
}

func (ss *ServerSuite) TestMethodIsRO() {
	assert.Equal(ss.T(), true, ss.srv.MethodIsRO("func_result"))
}

func (ss *ServerSuite) TestMethodIsROFalse() {
	assert.Equal(ss.T(), false, ss.srv.MethodIsRO("unknown"))
}

func (ss *ServerSuite) TestCallError() {

	ss.hook.Reset()

	var n *string
	a := map[string]interface{}{"code": n}

	tests := []struct {
		name   string
		method string
		args   map[string]interface{}
		err    string
	}{
		{name: "RequiredArgMissed", method: "func_result", err: "Required arg(s) missed (map[args:[code]])"},
		{name: "RequiredArgMissedRef", method: "func_result", args: a, err: "Required arg(s) missed (map[args:[code]])"},
		{name: "UnknownMethod", method: "unknown", err: "Method not found (map[name:unknown])"},
	}

	for _, tt := range tests {
		_, err := ss.srv.Call(ss.req, tt.method, tt.args)
		require.NotNil(ss.T(), err)
		assert.Equal(ss.T(), tt.err, err.Error())
	}

	// Two debug lines about required arg
	assert.Equal(ss.T(), 2, len(ss.hook.Entries))
	assert.Equal(ss.T(), logrus.DebugLevel, ss.hook.LastEntry().Level)
}

func (ss *ServerSuite) TestDBHIsNill() {
	_, err := New(ss.srv.config, ss.srv.log, nil)
	require.NotNil(ss.T(), err)
	assert.Equal(ss.T(), "dbh must be not nil", err.Error())
}

func (ss *ServerSuite) TearDownSuite() {
	fmt.Printf("exit\n")
}

func TestSuite(t *testing.T) {

	myTest := &ServerSuite{}
	suite.Run(t, myTest)
	/*
		myTest.hook.Reset()

		for _, e := range myTest.hook.Entries {
			fmt.Printf("ENT[%s]: %s\n", e.Level, e.Message)
		}
	*/
}
func (ss *ServerSuite) printLogs() {
	for _, e := range ss.hook.Entries {
		fmt.Printf("ENT[%s]: %s\n", e.Level, e.Message)
	}
}
