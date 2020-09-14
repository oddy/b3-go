package b3

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Varint API ---

func TestUvarintEncode(t *testing.T) {
	tests := []struct {
		input    uint64
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

func TestUvarintDecode(t *testing.T) {
	var tests = []struct {
		input []byte
		val   uint64 // setting type here makes it work. testify isn't good with
		index int    // untyped contants it seems.
		err   error
	}{
		{SBytes("32"), 50, 1, nil},
		{SBytes("f4 03"), 500, 2, nil},
		{SBytes("d0 86 03"), 50000, 3, nil},
		{SBytes("d0 86 83"), 0, 0, fmt.Errorf("uvarint > buffer")},
		{SBytes("ff ff ff ff ff ff ff ff ff 01"), 18_446_744_073_709_551_615, 10, nil},
		{SBytes("80 80 80 80 80 80 80 80 80 02"), 0, 0, fmt.Errorf("uvarint > uint64")},
	}
	// idiomatic method is to assert each return seperately.
	for _, test := range tests {
		val, index, err := DecodeUvarintInternal(test.input, 0)
		assert.Equal(t, err, test.err) // would use assert.Nil() if simpler.
		assert.Equal(t, index, test.index)
		assert.Equal(t, val, test.val)
	}
}


func TestSvarintEncode(t *testing.T) {
	tests := []struct {
		input    int64
		expected []byte
	}{
		{50, SBytes("64")},
		{-50, SBytes("63")},
		{123456789, SBytes("aa b4 de 75")},
		{-123456789, SBytes("a9 b4 de 75")},
	}
	for _, test := range tests {
		buf := make([]byte, 0)
		assert.Equal(t, test.expected, EncodeSvarint(buf, test.input))
	}
}

func TestSvarintDecode(t *testing.T) {
	var tests = []struct {
		input []byte
		val   int64  // setting type here makes it work. testify isn't good with go's normal somewhat-untyped constants
		index int    // untyped contants it seems.
		err   error
	}{
		{SBytes("64"), 50, 1, nil},
		{SBytes("63"), -50, 1, nil},
		{SBytes("aa b4 de 75"), 123456789, 4, nil},
		{SBytes("a9 b4 de 75"), -123456789, 4, nil},
		//{SBytes("ff ff ff ff ff ff ff ff ff 01"), 18_446_744_073_709_551_615, 10, nil},
		//{SBytes("80 80 80 80 80 80 80 80 80 02"), 0, 0, fmt.Errorf("uvarint > uint64")},
	}
	// idiomatic method is to assert each return seperately.
	for _, test := range tests {
		val, index, err := DecodeSvarintInternal(test.input, 0)
		assert.Equal(t, err, test.err) // would use assert.Nil() if simpler.
		assert.Equal(t, index, test.index)
		assert.Equal(t, val, test.val)
	}
}



