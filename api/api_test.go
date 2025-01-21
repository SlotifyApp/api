package api_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExample(t *testing.T) {
	assert := assert.New(t)

	a := "Hello"
	b := "Hello"

	assert.Equal(a, b, "The two words should be the same.")
}
