// +build db

package procapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ServerSuite struct {
	suite.Suite
	cfg  Config
	srv  *Service
	hook *test.Hook
	req  *http.Request
	conn DB //*sqlx.DB
	tx   Tx //*sqlx.Tx
	key  string
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

	ss.key = RandStringBytesRmndr(4)
	s := New(ss.cfg, log, nil).SetSchemaSuffix(ss.key)
	err = s.Open()
	require.NoError(ss.T(), err)

	db := s.DB()
	tx, err := db.Beginx()
	require.NoError(ss.T(), err)

	//	ss.cfg.DB.LogLevel = "debug" // we count log lines
	//	db, err := pgxpgcall.New(ss.cfg.DB, log)
	//	require.NoError(ss.T(), err)
	//	assert.Equal(ss.T(), "Added DB connection", ss.hook.LastEntry().Message)

	ss.req = httptest.NewRequest("GET", "http://example.com/foo", nil)
	//    w := httptest.NewRecorder()

	aliases := map[string]string{}
	for _, schema := range []string{"poma", "rpc", "rpc_testing"} {
		aliases[schema] = schema + "_" + ss.key
		err = loadPath(tx, schema, aliases)
		require.NoError(ss.T(), err)
	}

	err = s.LoadMethodsTx(tx)
	require.NoError(ss.T(), err)

	ss.srv = s
	ss.tx = tx
	ss.conn = db
}

func (ss *ServerSuite) TearDownSuite() {
	fmt.Printf("exit\n")
	ss.tx.Rollback()
	ss.conn.Close()
	/*
		if err != nil {
			pqe, ok := err.(*pq.Error)
			if ok != true {
				log.Fatal(err)
			}
			log.Fatalf("PG ERROR %s: %s\n", pqe.Code, err)
		}
	*/
}

func (ss *ServerSuite) TestCall() {

	ss.hook.Reset()

	tx := ss.tx
	//log := ss.srv.log

	//	Methods, err = Load(tx, key)
	checkTestUpdate("methods", ss.srv.methods)

	rv, err := ss.srv.CallTx(tx, "test_args", map[string]interface{}{"name": "xx"})
	require.NoError(ss.T(), err)

	assert.Equal(ss.T(), "xx", rv)

	args := map[string]interface{}{}
	helperLoadJSON(ss.T(), "test_types_args.json", &args)

	rv, err = ss.srv.CallTx(tx, "test_types", args)
	require.NoError(ss.T(), err)
	rv0 := rv.([]interface{})[0]
	checkTestUpdate("test_types_args", rv0)
	checkTestUpdate("test_types_rv", rv)

	rvWant := []interface{}{} //map[string]interface{}{}
	helperLoadJSON(ss.T(), "test_types_rv.json", &rvWant)

	data, err := json.Marshal(rv)
	require.NoError(ss.T(), err)

	rvGot := []interface{}{} //map[string]interface{}{}
	err = json.Unmarshal(data, &rvGot)
	require.NoError(ss.T(), err)

	assert.Equal(ss.T(), rvWant, rvGot)

	// Now we want another TZ
	_, err = tx.Exec("set timezone = 'Europe/Berlin'")
	require.NoError(ss.T(), err)
	rv, err = ss.srv.CallTx(tx, "test_types", args)
	require.NoError(ss.T(), err)
	rv1 := rv.([]interface{})[0]
	cet := rv1.(map[string]interface{})["ttimestamptz"]

	tstz, _ := time.Parse("02.01.2006 15:04:05.00 Z0700 MST", "17.12.1997 13:37:16.10 +0100 CET")
	assert.Equal(ss.T(), tstz.String(), cet.(time.Time).String())
}
