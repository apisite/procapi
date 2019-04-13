package procapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"

	"testing"

	"github.com/stretchr/testify/require"
)

func helperLoadJSON(t *testing.T, name string, data interface{}) {
	path := filepath.Join("testdata", name+".json") // relative path
	bytes, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	err = json.Unmarshal(bytes, &data)
	require.NoError(t, err)
}

const TestUpdateEnv = "TEST_UPDATE"
const TestUpdateDir = "testdata-new"

func helperCheckTestUpdate(file string, data interface{}) {
	if os.Getenv(TestUpdateEnv) == "" {
		return
	}
	if _, err := os.Stat(TestUpdateDir); os.IsNotExist(err) {
		err = os.Mkdir(TestUpdateDir, os.ModePerm)
		check(err)
	}
	p, err := ioutil.TempFile(TestUpdateDir, file+".*.json")
	check(err)
	fmt.Printf("*** Writing %s\n", p.Name())
	out, err := json.MarshalIndent(data, "", "  ")
	check(err)
	_, err = p.WriteString(string(out) + "\n") //ioutil.WriteFile(p, out, os.FileMode(mode))
	check(err)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// https://stackoverflow.com/a/31832326/5199825
const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

func RandStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
