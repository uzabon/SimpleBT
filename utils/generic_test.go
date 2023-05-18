package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStr(t *testing.T) {
	str := "123"
	res := Bytes(str)
	assert.Equal(t, []byte(str), res)
	assert.Equal(t, str, Str(res))
}
