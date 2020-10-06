package b3

import (
	"fmt"
	"math/bits"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Varint API ---

func TestUvarintEncode(t *testing.T) {
	tests := []struct {
		input    int
		expected []byte
	}{
		{50, SBytes("32")},
		{500, SBytes("f4 03")},
		{50000, SBytes("d0 86 03")},
		{0, SBytes("00")},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, EncodeUvarint(test.input))
	}
}

// There is bits.UintSize (in bits), and unsafe.Sizeof() (in bytes)

func TestEnsure64Bit(t *testing.T) {
	assert.Equal(t, 64, bits.UintSize, "!!! Remember to set GOARCH=amd64 !!!")				// we can use math/bits
	// var y int
	// assert.Equal(t, 8, int(unsafe.Sizeof(y)), "!!! Remember to set GOARCH=amd64 !!!")	// or unsafe.Sizeof
}

func TestUvarintDecode(t *testing.T) {
	var tests = []struct {
		input []byte
		val   int    // setting type here makes it work. testify isn't good with
		index int    // untyped contants it seems.
		err   error
	}{
		{SBytes("32"), 50, 1, nil},
		{SBytes("f4 03"), 500, 2, nil},
		{SBytes("d0 86 03"), 50000, 3, nil},
		{SBytes("d0 86 83"), 0, 0, fmt.Errorf("uvarint > buffer")},
		{SBytes("ff ff ff ff ff ff ff ff 7f"),    9_223_372_036_854_775_807, 9, nil},
		{SBytes("80 80 80 80 80 80 80 80 80 01"), 0, 0, fmt.Errorf("uvarint > int64")},
		// {SBytes("ff ff ff ff ff ff ff ff ff 01"), 18_446_744_073_709_551_615, 10, nil},	// todo: only if we go back to uint64
		// {SBytes("80 80 80 80 80 80 80 80 80 02"), 0, 0, fmt.Errorf("uvarint > uint64")}, // todo: only if we go back to uint64
	}
	// idiomatic method is to assert each return seperately.
	for _, test := range tests {
		val, index, err := DecodeUvarint(test.input, 0)
		assert.Equal(t, err, test.err) // would use assert.Nil() if simpler.
		assert.Equal(t, index, test.index)
		assert.Equal(t, val, test.val)
	}
}


func TestSvarintEncode(t *testing.T) {
	tests := []struct {
		input    int
		expected []byte
	}{
		{50, SBytes("64")},
		{-50, SBytes("63")},
		{123456789, SBytes("aa b4 de 75")},
		{-123456789, SBytes("a9 b4 de 75")},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, EncodeSvarint(test.input))
	}
}

func TestSvarintDecode(t *testing.T) {
	var tests = []struct {
		input []byte
		val   int    // setting type here makes it work. testify isn't good with go's normal somewhat-untyped constants
		index int    // untyped contants it seems.
		err   error
	}{
		{SBytes("64"), 50, 1, nil},
		{SBytes("63"), -50, 1, nil},
		{SBytes("aa b4 de 75"), 123456789, 4, nil},
		{SBytes("a9 b4 de 75"), -123456789, 4, nil},
		//{SBytes("ff ff ff ff ff ff ff ff ff 01"), 18_446_744_073_709_551_615, 10, nil},		// todo: error testing
		//{SBytes("80 80 80 80 80 80 80 80 80 02"), 0, 0, fmt.Errorf("uvarint > uint64")},
	}
	// idiomatic method is to assert each return seperately.
	for _, test := range tests {
		val, index, err := DecodeSvarint(test.input, 0)
		assert.Equal(t, err, test.err) // would use assert.Nil() if simpler.
		assert.Equal(t, index, test.index)
		assert.Equal(t, val, test.val)
	}
}



