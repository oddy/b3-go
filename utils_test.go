package b3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSBytes(t *testing.T) {
	foo := "0a 0a 40 40 64 64"
	expectedFoo := []byte{0x0a, 0x0a, 0x40, 0x40, 0x64, 0x64}
	assert.Equal(t, expectedFoo, SBytes(foo))

	bar := "64 65 66 67 68 69 70 71 72 73 74 75 76 77"
	expectedBar := []byte{0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77}
	assert.Equal(t, expectedBar, SBytes(bar))
}
