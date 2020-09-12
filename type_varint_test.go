package b3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// assert parameters tend to be t, then EXPECTED, then ACTUAL.

func TestEncodeUvarint(t *testing.T) {
	var tests = []struct {
		input    uint
		expected []byte
	}{
		{50, SBytes("32")},
		{500, SBytes("f4 03")},
		{50000, SBytes("d0 86 03")},
		{0, SBytes("00")},
	}
	for _, test := range tests {
		buf := make([]byte, 0)
		assert.Equal(t, test.expected, EncodeUvarint(buf, test.input))
	}
}

func TestEncodeSvarint(t *testing.T) {
	var tests = []struct {
		input    int
		expected []byte
	}{
		{50,  SBytes("64")},
		{-50, SBytes("63")},
		{123456789,  SBytes("aa b4 de 75")},
		{-123456789, SBytes("a9 b4 de 75")},
	}
	for _, test := range tests {
		buf := make([]byte, 0)
		assert.Equal(t, test.expected, EncodeSvarint(buf, test.input))
	}

}

