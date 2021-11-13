package procapi

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
//	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"time"

	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/pgmig/pgmig"

	"github.com/jackc/pgx/v4"
)

type ServerSuite struct {
	suite.Suite
	cfg  Config
	srv  *Service
	hook *test.Hook
	req  *http.Request
	db   *pgx.Conn
	tx   pgx.Tx
	wg   sync.WaitGroup
	mig  *pgmig.Migrator
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

	ctx := context.Background()
	var cfgMig pgmig.Config
	p = flags.NewParser(&cfgMig, flags.Default|flags.IgnoreUnknown)
	_, err = p.Parse()
	require.NoError(ss.T(), err)
	ss.mig = pgmig.New(cfgMig, log, nil, "testdata")

	ss.db, err = ss.mig.Connect(os.Getenv("TEST_DATABASE"))
	require.NoError(ss.T(), err)

	ss.tx, err = ss.db.Begin(ctx)
	require.NoError(ss.T(), err)

	ss.wg.Add(1)
	go ss.mig.PrintMessages(&ss.wg)
	_, err = ss.mig.Run(ss.tx, "init", []string{"pgmig", "rpc", "rpc_testing"})
	require.NoError(ss.T(), err)

	s := New(ss.cfg, log, ss.db)

	ss.req = httptest.NewRequest("GET", "http://example.com/foo", nil)
	//    w := httptest.NewRecorder()

	err = s.LoadMethodsTx(ss.tx)
	require.NoError(ss.T(), err)

	helperCheckTestUpdate("methods", s.methods)

	ss.srv = s

}

func (ss *ServerSuite) TearDownSuite() {
	fmt.Printf("exit\n")
	ss.tx.Rollback(context.Background())
//	ss.tx.Commit(context.Background())
	ss.db.Close(context.Background())

	close(ss.mig.MessageChan)
	ss.wg.Wait()

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
	var str1 interface{} = "xx"
	rv, err := ss.srv.CallTx(tx, "test_args", map[string]interface{}{"name": &str1})
	require.NoError(ss.T(), err)
	assert.Equal(ss.T(), &str1, rv)

	args := map[string]interface{}{}
	helperLoadJSON(ss.T(), "test_types_args", &args)
	args["tint8"] = int8(args["tint8"].(float64))
	args["tint4"] = int(args["tint4"].(float64))
	args["tint2"] = int(args["tint2"].(float64))

	rv, err = ss.srv.CallTx(tx, "test_types", args)
	require.NoError(ss.T(), err)
	require.NotNil(ss.T(), rv)

	rv0 := rv.([]interface{})[0]
	rv01 := rv0.(map[string]interface{})

	rv01["tfloat4"] = math.Round(float64(rv01["tfloat4"].(float32))*1000) / 1000

	helperCheckTestUpdate("test_types_args", rv01)
	helperCheckTestUpdate("test_types_rv", rv)

	rvWant := []interface{}{}
	helperLoadJSON(ss.T(), "test_types_rv", &rvWant)

	data, err := json.Marshal(rv)
	require.NoError(ss.T(), err)

	rvGot := []interface{}{}
	err = json.Unmarshal(data, &rvGot)
	require.NoError(ss.T(), err)

	assert.Equal(ss.T(), rvWant, rvGot)

	ctx := context.Background()
	// Now we want another TZ
	_, err = tx.Exec(ctx, "set timezone = 'Europe/Berlin'")
	require.NoError(ss.T(), err)
	rv, err = ss.srv.CallTx(tx, "test_types", args)
	require.NoError(ss.T(), err)
	rv1 := rv.([]interface{})[0]
	cet := rv1.(map[string]interface{})["ttimestamptz"]

	tstz, _ := time.Parse("02.01.2006 15:04:05.00 Z0700 MST", "17.12.1997 13:37:16.10 +0100 CET")
	assert.Equal(ss.T(), tstz.String(), cet.(time.Time).String())
}

func (ss *ServerSuite) TestCallVoid() {

	ss.hook.Reset()

	tx := ss.tx

	_, err := ss.srv.CallTx(tx, "test_void", map[string]interface{}{"code": "arg"})
	require.NoError(ss.T(), err)
}
