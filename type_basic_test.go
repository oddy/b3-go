package b3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseBoolEnc(t *testing.T) {
	assert.Equal(t, SBytes("01"), EncodeBool(true))
	assert.Equal(t, SBytes(""),   EncodeBool(false))
}

func TestBaseUtf8Enc(t *testing.T) {
	assert.Equal(t, SBytes("68 65 6c 6c 6f 20 77 6f 72 6c 64"), EncodeUtf8("hello world"))
	assert.Equal(t, SBytes("d0 92 d0 b8 d0 b0 d0 b3 d1 80 d0 b0"), EncodeUtf8("Виагра"))										// Viagra OWEN
	assert.Equal(t, SBytes("e2 9c 88 e2 9c 89 f0 9f 9a 80 f0 9f 9a b8 f0 9f 9a bc f0 9f 9a bd"), EncodeUtf8("✈✉🚀🚸🚼🚽"))		// SMP
	assert.Equal(t, SBytes(""), EncodeUtf8(""))
}

func TestBaseInt64Enc(t *testing.T) {
	assert.Equal(t, SBytes("15 cd 5b 07 00 00 00 00"), EncodeInt64(123456789))
	assert.Equal(t, SBytes("eb 32 a4 f8 ff ff ff ff"), EncodeInt64(-123456789))
	assert.Equal(t, SBytes(""), EncodeInt64(0))
}

func TestBaseFloat64Enc(t *testing.T) {
	assert.Equal(t, SBytes("a1 f8 31 e6 d6 1c c8 40"), EncodeFloat64(12345.6789))
	assert.Equal(t, SBytes(""), EncodeFloat64(0.0))
}

//
