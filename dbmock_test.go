// +build !db

package procapi

//go:generate mockgen -destination=generated_mock_test.go -package procapi -source=procapi.go DB,Tx,Rows,Result

import (
	"net/http"
	"net/http/httptest"
	"testing"

	//	"github.com/jmoiron/sqlx"

	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mitchellh/mapstructure"

	"github.com/golang/mock/gomock"

	"github.com/apisite/procapi/pgtype"
)

type ServerSuite struct {
	suite.Suite
	cfg  Config
	srv  *Service
	hook *test.Hook
	req  *http.Request
	db   *MockDB
	tx   *MockTx
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
	s := New(ss.cfg, log, m).SetMarshaller(pgtype.New())

	s.LoadMethods()
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

	tx := NewMockTx(ctrl)

	m.EXPECT().Beginx().
		Return(tx, nil)

	indexResp := NewMockRows(ctrl)
	tx.EXPECT().Queryx("select * from rpc.index($1)", nil).
		Return(indexResp, nil) //indexRows, nil)
	expectStructTable(ss.T(), indexResp, indexRows, &Method{})

	for _, method := range indexRows {
		code := method["code"].(string)
		argsResp := NewMockRows(ctrl)
		a := []interface{}{code}
		tx.EXPECT().Queryx("select * from rpc.func_args($1)", a).
			Return(argsResp, nil) //allArgs[code], nil)
		expectStructTable(ss.T(), argsResp, allArgs[code], &InDef{})

		resResp := NewMockRows(ctrl)
		tx.EXPECT().Queryx("select * from rpc.func_result($1)", a). //[]interface{}{code}).
										Return(resResp, nil) //allResult[code], nil)
		expectStructTable(ss.T(), resResp, allResult[code], &OutDef{})
	}
	tx.EXPECT().Rollback().Return(nil)

}

func (ss *ServerSuite) TestCall() {

	var allResult map[string][]map[string]interface{}
	//var allResult map[string][]interface{} //map[string]interface{}
	helperLoadJSON(ss.T(), "result.json", &allResult)
	var allFields map[string][]string
	helperLoadJSON(ss.T(), "fields.json", &allFields)

	ctrl := gomock.NewController(ss.T())
	defer ctrl.Finish()
	m := ss.db

	ss.hook.Reset()
	tx := NewMockTx(ctrl)
	m.EXPECT().Beginx().
		Return(tx, nil)

	tests := []struct {
		name   string
		method string
		args   map[string]interface{}
		res    []map[string]interface{}
		//		err    string
	}{
		{name: "Res", method: "func_result", args: map[string]interface{}{"code": "index"}, res: allResult["index"]},
	}
	// If tests will grow - move the following inside test loop
	indexResp := NewMockRows(ctrl)
	tx.EXPECT().Queryx("select arg, type, anno from rpc.func_result(a_code := $1)", []interface{}{"index"}).
		Return(indexResp, nil) //allResult["index"]
	expectMapTable(ss.T(), indexResp, allResult["index"])
	tx.EXPECT().Commit().Return(nil)

	for _, tt := range tests {
		rv, err := ss.srv.Call(ss.req, tt.method, tt.args)
		require.NoError(ss.T(), err)
		v, ok := rv.([]interface{})
		assert.True(ss.T(), ok)
		v1 := []map[string]interface{}{}
		for _, s := range v {
			z, ok := s.(map[string]interface{})
			assert.True(ss.T(), ok)
			v1 = append(v1, z)
		}

		assert.Equal(ss.T(), tt.res, v1)
	}

	// Two debug lines about required arg + SQL
	ss.printLogs()                                // show logs
	assert.Equal(ss.T(), 3, len(ss.hook.Entries)) // count logs

}

func expectStructTable(t *testing.T, rows *MockRows, maps []map[string]interface{}, rowStruct interface{}) {
	for _, m := range maps {
		rows.EXPECT().Next().
			Return(true)
		rows.EXPECT().StructScan(rowStruct).DoAndReturn(
			func(x map[string]interface{}) func(v interface{}) error {
				return func(v interface{}) error {
					err := decode(x, v)
					require.NoError(t, err)
					return nil
				}
			}(m))
	}
	rows.EXPECT().Next().
		Return(false)
	rows.EXPECT().Close()

}
func expectMapTable(t *testing.T, rows *MockRows, maps []map[string]interface{}) {
	for _, m := range maps {
		rows.EXPECT().Next().
			Return(true)
		z := map[string]interface{}{}
		rows.EXPECT().MapScan(z).DoAndReturn(
			func(x map[string]interface{}) func(rv map[string]interface{}) error {
				return func(rv map[string]interface{}) error {
					for k, v := range x {
						rv[k] = v
					}
					return nil
				}
			}(m))
	}
	rows.EXPECT().Next().
		Return(false)
	rows.EXPECT().Close()

}

// decode fills struct from map using ithub.com/mitchellh/mapstructure
func decode(input interface{}, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   output,
		TagName:  "db",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
