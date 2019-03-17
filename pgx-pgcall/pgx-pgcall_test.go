package pgxpgcall

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRowsAffected(t *testing.T) {

	result := Result{CommandTag: "rows: 72"}
	var rows int64 = 72
	rv, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, rows, rv)
}
