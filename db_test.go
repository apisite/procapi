// +build db

package pgcall

import (
	"net/http"
	"net/http/httptest"

	"encoding/json"
	"net"
	"time"

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
	Call Config           `group:"pgcall Options" namespace:"pgcall"`
	DB   pgxpgcall.Config `group:"pgx-pgcall Options" namespace:"db" env-namespace:"DB"`
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
	_, err := p.Parse()
	require.NoError(ss.T(), err)

	l, hook := test.NewNullLogger()
	ss.hook = hook
	l.SetLevel(logrus.DebugLevel)
	log := mapper.NewLogger(l)

	hook.Reset()
	ss.cfg.DB.LogLevel = "debug" // we count log lines
	db, err := pgxpgcall.New(ss.cfg.DB, log)
	require.NoError(ss.T(), err)
	assert.Equal(ss.T(), "Added DB connection", ss.hook.LastEntry().Message)

	ss.req = httptest.NewRequest("GET", "http://example.com/foo", nil)
	//    w := httptest.NewRecorder()

	s, err := New(ss.cfg.Call, log, db)
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

	// 3 debug lines about required arg (IN args, Use, Sql, Query)
	assert.Equal(ss.T(), 4, len(ss.hook.Entries))
}

func (ss *ServerSuite) TestCallArgs() {

	ss.hook.Reset()

	var s *string
	tests := []struct {
		name    string
		method  string
		args    map[string]interface{}
		res     interface{}
		wantNil bool
	}{
		{name: "Regular", method: "test_args", args: map[string]interface{}{"name": "index"}, res: "index"},
		{name: "Null", method: "test_args", args: map[string]interface{}{"name": s}, wantNil: true},
		{name: "Default", method: "test_args", args: map[string]interface{}{}, res: "def"},
	}

	for _, tt := range tests {
		rv, err := ss.srv.Call(ss.req, tt.method, tt.args)
		require.NoError(ss.T(), err)
		//		checkTestUpdate(tt.name, rv)
		if tt.wantNil {
			assert.Nil(ss.T(), rv)
		} else {
			assert.Equal(ss.T(), tt.res, rv)

		}
	}
	ss.printLogs()                                 // show logs
	assert.Equal(ss.T(), 13, len(ss.hook.Entries)) // count logs
}

func (ss *ServerSuite) TestCallTypes() {

	ss.hook.Reset()

	_, ipv4Net, _ := net.ParseCIDR("192.0.2.1/24")
	dt, _ := time.Parse("2006-01-02", "2011-01-19")
	dt2, _ := time.Parse("2006-01-02", "2019-03-31")

	ts, _ := time.Parse("02.01.2006 15:04:05.00", "17.12.1997 15:37:16.10")
	ts2, _ := time.Parse("02.01.2006 15:04:05", "17.12.1997 15:37:16")

	tstz, _ := time.Parse("02.01.2006 15:04:05.00 MST", "17.12.1997 12:37:16.10 UTC")
	tstz2, _ := time.Parse("02.01.2006 15:04 MST", "17.12.1997 16:00 UTC")

	// Will receive timestamps for this location
	// ENV{TZ} should be used to set timezone (TODO: How is this works?)
	loc := time.Now().Location()

	args := map[string]interface{}{
		"tbool":        bool(true),
		"tchar":        string("z"),
		"tdate":        dt,
		"tfloat4":      float32(12.34),
		"tfloat8":      float64(3456.7890),
		"tinet":        ipv4Net, //"127.0.1.2/8",
		"tint2":        int16(2),
		"tint4":        int(4),
		"tint8":        int64(8),
		"tinterval":    "10s", //time.Duration(10) * time.Second,
		"tjson":        json.RawMessage(`{"precomputed": true, "b":2}`),
		"tjsonb":       json.RawMessage(`{"precomputed": true, "b":2}`),
		"tmoney":       "5678.9012", //float32(5678.9012),
		"tnumeric":     float32(7890.1234),
		"ttext":        `{"precomputed": true, "b":2}`,
		"ttime":        "23:59:10", //tm,
		"ttimestamp":   ts,
		"ttimestamptz": tstz.In(loc),
		"aint4":        []int32{1, 2, 3},
		"atext":        []string{`{"b":2}`, `{"c":3}`},
	}

	want := []map[string]interface{}{
		{
			"aint4":        []int32{1, 2, 3},
			"atext":        []string{`{"b":2}`, `{"c":3}`},
			"id":           int32(1),
			"tbool":        true,
			"tchar":        "z",
			"tdate":        dt,
			"tfloat4":      float32(12.34),
			"tfloat8":      3456.789,
			"tinet":        ipv4Net,
			"tint2":        int16(2),
			"tint4":        int32(4),
			"tint8":        int64(8),
			"tinterval":    time.Duration(10) * time.Second,
			"tjson":        map[string]interface{}{"b": float64(2), "precomputed": true},
			"tjsonb":       map[string]interface{}{"b": float64(2), "precomputed": true},
			"tmoney":       "$5,678.90",
			"tnumeric":     float32(7890.1234),
			"ttext":        `{"precomputed": true, "b":2}`,
			"ttime":        "23:59:10",
			"ttimestamp":   ts,
			"ttimestamptz": tstz.In(loc),
		},
		{
			"aint4":        []int32{9, 8, 7},
			"atext":        []string{"zyx1", "zyx2"},
			"id":           int32(2),
			"tbool":        false,
			"tchar":        "x",
			"tdate":        dt2,
			"tfloat4":      float32(4.113333),
			"tfloat8":      float64(1152.2630000000001),
			"tinet":        ipv4Net,
			"tint2":        int16(1),
			"tint4":        int32(2),
			"tint8":        int64(4),
			"tinterval":    time.Duration(3610) * time.Second, // 1h0m10s
			"tjson":        map[string]interface{}{"b": float64(2), "precomputed": true},
			"tjsonb":       map[string]interface{}{"b": float64(2), "precomputed": true},
			"tmoney":       "$5,678.90",
			"tnumeric":     float32(7890.1234),
			"ttext":        `{"precomputed": true, "b":2}{"precomputed": true, "b":2}`,
			"ttime":        "23:55:10.5",
			"ttimestamp":   ts2,
			"ttimestamptz": tstz2.In(loc),
		},
		{
			"id": int32(3),
		},
	}

	rv, err := ss.srv.Call(ss.req, "test_types", args)
	require.NoError(ss.T(), err)

	checkTestUpdate("test_types", rv)
	assert.Equal(ss.T(), want, rv)

}
