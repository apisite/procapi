package procapi

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"path/filepath"

	"testing"

	"github.com/stretchr/testify/require"
)

func helperLoadJSON(t *testing.T, name string, data interface{}) {
	path := filepath.Join("testdata", name) // relative path

	/* TODO
	if *update {
		bytes, err := json.Marshal(data)
		require.NoError(t, err)
		ioutil.WriteFile(path, bytes, 0644)
		return
	}
	*/
	bytes, err := ioutil.ReadFile(path)
	require.NoError(t, err)

	err = json.Unmarshal(bytes, &data)
	require.NoError(t, err)

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
