package eris

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := New("test error")
	wrapErr := Wrap(err, "additional context")
	assert.Equal(t, wrapErr.Error(), "additional context: test error")
}
