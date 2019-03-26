package ginpgcall

//go:generate mockgen -destination=generated_mock_test.go -package ginpgcall github.com/apisite/pgcall/gin-pgcall Caller

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/golang/mock/gomock"
)

type ServerSuite struct {
	suite.Suite
	srv  *Server
	hook *test.Hook
	mock *MockCaller
}

func (ss *ServerSuite) SetupSuite() {

	l, hook := test.NewNullLogger()
	ss.hook = hook
	l.SetLevel(logrus.DebugLevel)
	log := mapper.NewLogger(l)
	hook.Reset()

	ctrl := gomock.NewController(ss.T())
	defer ctrl.Finish()

	ss.mock = NewMockCaller(ctrl)
	ss.srv = NewServer(log, ss.mock)

}

func TestSuite(t *testing.T) {

	myTest := &ServerSuite{}
	suite.Run(t, myTest)

	myTest.hook.Reset()

	for _, e := range myTest.hook.Entries {
		fmt.Printf("ENT[%s]: %s\n", e.Level, e.Message)
	}

}

func (ss *ServerSuite) TestHandler() {
	r := gin.Default()
	allFuncs := template.FuncMap{}

	s := ss.srv
	s.SetProtoFuncs(allFuncs)
	s.Route("/rpc", r)

	m := ss.mock

	req, _ := http.NewRequest("GET", "/rpc/index", nil)
	resp := httptest.NewRecorder()

	m.EXPECT().Call(req, "index", map[string]interface{}{}).
		Return("bar", nil)

	r.ServeHTTP(resp, req)
	assert.Equal(ss.T(), resp.Body.String(), `"bar"`)

	data := strings.NewReader("{}")
	req, _ = http.NewRequest("POST", "/rpc/index", data)
	resp = httptest.NewRecorder()
	m.EXPECT().Call(req, "index", map[string]interface{}{}).
		Return("bar", nil)

	r.ServeHTTP(resp, req)
	assert.Equal(ss.T(), resp.Body.String(), `"bar"`)

}

func (ss *ServerSuite) TestFunc() {
	// First we create a FuncMap with which to register the function.
	funcMap := template.FuncMap{
		"title": strings.Title,
	}
	s := ss.srv
	s.SetProtoFuncs(funcMap)

	/*
	   	const templateText = `
	   Input: {{printf "%q" .}}
	   Output 0: {{title .}}
	   Output 1: {{title . | printf "%q"}}
	   Output 2: {{printf "%q" . | title}}
	   `
	*/
	const templateText = `
{{ makeSlice "a" "b" | printf "%v"}}
{{ makeMap "a" 1 "b" 2 | printf "%v"}}
`
	// Create a template, add the function map, and parse the text.
	tmpl, err := template.New("titleTest").Funcs(funcMap).Parse(templateText)
	require.NoError(ss.T(), err)

	// Run the template to verify the output.
	data := bytes.NewBuffer([]byte{})

	err = tmpl.Execute(data, "the go programming language")
	require.NoError(ss.T(), err)

	/*
	   	out := `
	   Input: &#34;the go programming language&#34;
	   Output 0: The Go Programming Language
	   Output 1: &#34;The Go Programming Language&#34;
	   Output 2: &#34;The Go Programming Language&#34;
	   `
	*/
	out := `
[a b]
&amp;map[a:1 b:2]
`
	assert.Equal(ss.T(), data.String(), out)

}
