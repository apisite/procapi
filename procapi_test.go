package procapi

import (
	"fmt"
	"testing"

	//	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func (ss *ServerSuite) TestMethod() {
	var mGot, mWant []Method
	cfg := ss.cfg
	for _, name := range []string{cfg.IndexFunc, cfg.InDefFunc, cfg.OutDefFunc} {
		if m, ok := ss.srv.Method(name); ok {
			m.Class = "rpc"
			if name == cfg.IndexFunc {
				var s = "rpc.func_def"
				m.Result = &s
			}
			mGot = append(mGot, m)
		}
	}

	checkTestUpdate("methods.golden.json", mGot)
	helperLoadJSON(ss.T(), "methods.golden.json", &mWant)
	assert.Equal(ss.T(), mWant, mGot)
}

func (ss *ServerSuite) TestMethodIsRO() {
	m, ok := ss.srv.Method(ss.cfg.OutDefFunc)
	assert.True(ss.T(), ok && m.IsRO)
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
		{name: "RequiredArgMissed", method: ss.cfg.OutDefFunc, err: "Required arg(s) missed (map[args:[code]])"},
		{name: "RequiredArgMissedRef", method: ss.cfg.OutDefFunc, args: a, err: "Required arg(s) missed (map[args:[code]])"},
		{name: "UnknownMethod", method: "unknown", err: "Method not found (map[name:unknown])"},
	}

	for _, tt := range tests {
		_, err := ss.srv.CallTx(ss.tx, tt.method, tt.args)
		require.NotNil(ss.T(), err)
		assert.Equal(ss.T(), tt.err, err.Error())
	}

	// Two debug lines about required arg
	//assert.Equal(ss.T(), 2, len(ss.hook.Entries))
	//	assert.Equal(ss.T(), logrus.DebugLevel, ss.hook.LastEntry().Level)
}

func (ss *ServerSuite) TestDBHIsNill() {
	db := New(ss.srv.config, ss.srv.log, nil)
	err := db.LoadMethods()
	assert.Equal(ss.T(), "dbh must be not nil", err.Error())
	_, err = db.Call(ss.req, "any", nil)
	assert.Equal(ss.T(), "dbh must be not nil", err.Error())
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
