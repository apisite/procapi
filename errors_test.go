package procapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnknown(t *testing.T) {
	c := callError{code: errUnknown}
	cu := callError{code: 99}
	assert.Equal(t, c.Code(), cu.Code(), "Unknown code Code eq errUnknown")
	assert.Equal(t, c.Message(), cu.Message(), "Unknown code Message eq errUnknown")
}
