package eris_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/morningvera/eris"
)

func TestError(t *testing.T) {
	err := eris.New("test error")
	wrapErr := eris.Wrap(err, "additional context")
	assert.Equal(t, wrapErr.Error(), "additional context: test error")
}
