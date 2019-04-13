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

	helperCheckTestUpdate("methods.golden", mGot)
	helperLoadJSON(ss.T(), "methods.golden", &mWant)
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
	aErr := map[string]interface{}{"args": []string{"code"}}
	tests := []struct {
		name           string
		method         string
		args           map[string]interface{}
		err            string
		rvCode         string
		rvData         map[string]interface{}
		rvIsNotFound   bool
		rvIsBadRequest bool
	}{
		{name: "RequiredArgsMissed", method: ss.cfg.OutDefFunc,
			err:            "Required arg(s) missed (map[args:[code]])",
			rvIsBadRequest: true, rvData: aErr},
		{name: "RequiredArgsMissed", method: ss.cfg.OutDefFunc, args: a,
			err:            "Required arg(s) missed (map[args:[code]])",
			rvIsBadRequest: true, rvData: aErr},
		{name: "MethodNotFound", method: "unknown", err: "Method not found (map[name:unknown])", rvIsNotFound: true,
			rvData: map[string]interface{}{"name": "unknown"}},
	}

	for _, tt := range tests {
		_, err := ss.srv.CallTx(ss.tx, tt.method, tt.args)
		require.NotNil(ss.T(), err)
		assert.Equal(ss.T(), tt.err, err.Error())
		cerr, ok := err.(*callError)
		assert.True(ss.T(), ok)
		if ok {
			assert.Equal(ss.T(), tt.name, cerr.Code())
			assert.Equal(ss.T(), tt.rvData, cerr.Data())
			assert.Equal(ss.T(), tt.rvIsNotFound, cerr.IsNotFound())
			assert.Equal(ss.T(), tt.rvIsBadRequest, cerr.IsBadRequest())
		}
	}

	// Two debug lines about required arg
	assert.Equal(ss.T(), 2, len(ss.hook.Entries))
	//	assert.Equal(ss.T(), logrus.DebugLevel, ss.hook.LastEntry().Level)
}

func (ss *ServerSuite) TestDBHIsNill() {
	db := New(ss.srv.config, ss.srv.log, nil)

	err := db.LoadMethods()
	require.NotNil(ss.T(), err)
	cerr, ok := err.(*callError)
	assert.True(ss.T(), ok)
	assert.Equal(ss.T(), "NilDB", cerr.Code())

	_, err = db.Call(ss.req, "any", nil)
	require.NotNil(ss.T(), err)
	cerr, ok = err.(*callError)
	assert.True(ss.T(), ok)
	assert.Equal(ss.T(), "NilDB", cerr.Code())
}

func TestSuite(t *testing.T) {

	myTest := &ServerSuite{}
	suite.Run(t, myTest)

}
func (ss *ServerSuite) printLogs() {
	for _, e := range ss.hook.Entries {
		fmt.Printf("ENT[%s]: %s\n", e.Level, e.Message)
	}
}
