// +build !db

package pgcall

//go:generate mockgen -destination=generated_mock_test.go -package pgcall github.com/apisite/pgcall/pgiface Rows,Result,DB

import (
	"net/http"
	"net/http/httptest"

	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/golang/mock/gomock"
)

type ServerSuite struct {
	suite.Suite
	cfg  Config
	srv  *Server
	hook *test.Hook
	req  *http.Request
	db   *MockDB
}

func (ss *ServerSuite) SetupSuite() {

	// Fill config with default values
	p := flags.NewParser(&ss.cfg, flags.Default)
	_, err := p.ParseArgs([]string{})
	require.NoError(ss.T(), err)

	l, hook := test.NewNullLogger()
	ss.hook = hook
	l.SetLevel(logrus.DebugLevel)
	log := mapper.NewLogger(l)
	hook.Reset()

	ss.req = httptest.NewRequest("GET", "http://example.com/foo", nil)
	//    w := httptest.NewRecorder()

	ctrl := gomock.NewController(ss.T())
	defer ctrl.Finish()

	m := NewMockDB(ctrl)

	ss.prepServer(ctrl, m)

	s, err := New(ss.cfg, log, m)
	require.NoError(ss.T(), err)

	ss.srv = s
	ss.db = m

}

func (ss *ServerSuite) prepServer(ctrl *gomock.Controller, m *MockDB) {

	t := ss.T()

	var indexRows []map[string]interface{}
	helperLoadJSON(t, "index.json", &indexRows)

	var allArgs map[string][]map[string]interface{}
	helperLoadJSON(t, "args.json", &allArgs)

	var allResult map[string][]map[string]interface{}
	helperLoadJSON(t, "result.json", &allResult)

	var allFields map[string][]string
	helperLoadJSON(t, "fields.json", &allFields)

	indexResp := NewMockRows(ctrl)
	m.EXPECT().Query("select * from index()", []interface{}{}).
		Return(indexResp, nil)
	expectTable(indexResp, allFields["index"], indexRows)

	for _, method := range indexRows {
		code := method["code"].(string)

		argsResp := NewMockRows(ctrl)
		m.EXPECT().Query("select * from func_args($1)", []interface{}{code}).
			Return(argsResp, nil)
		expectTable(argsResp, allFields["args"], allArgs[code])

		resResp := NewMockRows(ctrl)
		m.EXPECT().Query("select * from func_result($1)", []interface{}{code}).
			Return(resResp, nil)
		expectTable(resResp, allFields["result"], allResult[code])

	}

}

func (ss *ServerSuite) TestCall() {

	var allResult map[string][]map[string]interface{}
	helperLoadJSON(ss.T(), "result.json", &allResult)
	var allFields map[string][]string
	helperLoadJSON(ss.T(), "fields.json", &allFields)

	ctrl := gomock.NewController(ss.T())
	defer ctrl.Finish()
	m := ss.db

	ss.hook.Reset()

	tests := []struct {
		name   string
		method string
		args   map[string]interface{}
		res    []map[string]interface{}
		err    string
	}{
		{name: "Res", method: "func_result", args: map[string]interface{}{"code": "index"}, res: allResult["index"]},
	}
	// If tests will grow - move the following inside test loop
	indexResp := NewMockRows(ctrl)
	m.EXPECT().Query("select arg, type, anno from pgfc_test.func_result(a_code := $1)", []interface{}{"index"}).
		Return(indexResp, nil)
	expectTable(indexResp, allFields["result"], allResult["index"])

	for _, tt := range tests {
		rv, err := ss.srv.Call(ss.req, tt.method, tt.args)
		require.NoError(ss.T(), err)
		assert.Equal(ss.T(), tt.res, rv)
	}

	// Two debug lines about required arg + SQL
	assert.Equal(ss.T(), 3, len(ss.hook.Entries))
	//assert.Equal(ss.T(), logrus.DebugLevel, ss.hook.LastEntry().Message)
}

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
