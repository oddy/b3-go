package b3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSBytes(t *testing.T) {
	foo := "0a 0a 40 40 64 64"
	assert.Equal(t, []byte("\x0a\x0a\x40\x40\x64\x64"), SBytes(foo))

	bar := "64 65 66 67 68 69 70 71 72 73 74 75 76 77"
	assert.Equal(t, []byte("\x64\x65\x66\x67\x68\x69\x70\x71\x72\x73\x74\x75\x76\x77"), SBytes(bar))
}
