package pgcall

import (
	"encoding/json"
	"io/ioutil"
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
