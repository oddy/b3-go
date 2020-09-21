package b3

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
 Item:
 [header BYTE] [15+ type# UVARINT] [key (see below)] [data len UVARINT]  [ data BYTES ]
 ---------------------------- item_header -----------------------------  --- codecs ---

 --- header byte ---
 +------------+------------+------------+------------+------------+------------+------------+------------+
 | is null    | has data   | key type   | key type   | data type  | data type  | data type  | data type  |
 +------------+------------+------------+------------+------------+------------+------------+------------+

 +------------+------------+
 | is null    | has data   |
 +------------+------------+
     1   x  (2)    Value is None/NULL/nil - data len & has data ignored
     0   0  (0)    Codec zero-value for given data type (0, "", 0.0 etc)
     0   1  (1)    Data len present, followed by codec'ed data bytes

 +------------+------------+
 | key type   | key type   |
 +------------+------------+
     0   0  (0)    no key
     0   1  (4)    UVARINT
     1   0  (8)    UTF8 bytes
     1   1  (c)    raw bytess
*/

// =====================================================================================================================
// = Item header

// Args:         dataType,  key,  isNull,  dataLen

// --- Header null & has-data bits ENcoder ---
func TestHeaderNullEnc(t *testing.T) {
	buf, err := EncodeHeader(0, nil, true, 0)
	assert.Nil(t, err)
	assert.Equal(t, SBytes("80"), buf) // isNull true
}

func TestHeaderHasdataEnc(t *testing.T) {
	buf, err := EncodeHeader(0, nil, false, 5)
	assert.Nil(t, err)
	assert.Equal(t, SBytes("40 05"), buf)					// has-data on, size follows
}

func TestHeaderZerovalEnc(t *testing.T) {
	buf, err := EncodeHeader(0, nil, false, 0)
	assert.Nil(t, err)
	assert.Equal(t, SBytes("00"), buf)					// not null but no data = compact zero-value mode
}

// Policy: Encoder: is_null supercedes any datalen info. If null is on, data_len forced to 0, has_data forced to false.

func TestHeaderHasdataButNullEnc(t *testing.T) {
	buf, err := EncodeHeader(0, nil, true, 5)
	assert.Nil(t, err)
	assert.Equal(t, SBytes("80"), buf)					// test that isNull supercedes dataLen
}

// --- Data len ---

func TestHeaderDatalenEnc(t *testing.T) {
	buf, err := EncodeHeader(5, nil, false, 5)
	assert.Nil(t, err)
	assert.Equal(t, SBytes("45 05"), buf)
	buf, err =  EncodeHeader(5, nil, false, 1500)
	assert.Nil(t, err)
	assert.Equal(t, SBytes("45 dc 0b"), buf)
}

// --- Ext data type numbers ---

func TestHeaderDatatypeEnc(t *testing.T) {
	tests := []struct {
		dataType int
		buf []byte
	}{
		{5,   SBytes("05")},
		{14,  SBytes("0e")},
		{15,  SBytes("0f 0f")},
		{16,  SBytes("0f 10")},
		{555, SBytes("0f ab 04")},
	}
	for _,test := range tests {
		buf, err := EncodeHeader(test.dataType, nil, false, 0)
		assert.Nil(t, err)
		assert.Equal(t, test.buf, buf)
	}
}

// --- Keys ---

func TestHeaderKeysEnc(t *testing.T) {
	tests := []struct {
		key interface{}
		buf []byte
	}{
		{nil, 			SBytes("00")},
		{4,   			SBytes("10 04")},
		{7777777, 		SBytes("10 f1 db da 03")},
		{"foo",			SBytes("20 03 66 6f 6f")},
		{"Виагра",  	SBytes("20 0c d0 92 d0 b8 d0 b0 d0 b3 d1 80 d0 b0")},
		{[]byte("foo"), SBytes("30 03 66 6f 6f")},
	}
	for _, test := range tests {
		buf, err := EncodeHeader(0, test.key, false, 0)
		assert.Nil(t, err)
		assert.Equal(t, test.buf, buf)
	}
}


// --- Kitchen sink ---

func TestHeaderAllEnc(t *testing.T) {
	buf, err := EncodeHeader(555, "foo", false, 1500)
	assert.Nil(t, err)
	exBuf := SBytes("6f ab 04 03 66 6f 6f dc 0b")
    //               --                              control: null=no  data=yes  key=1,0 (UTF8)  data_type=extended (1,1,1,1)
    //                  -----                        ext type uvarint (555)
    //                        --                     len of utf8 key (3 bytes)
    //                           --------            utf8 key "foo"
    //                                    -----      data len (1500)
    assert.Equal(t, exBuf, buf)
}



// =====================================================================================================================
// = Item header keys

func TestKeytypeEnc(t *testing.T) {
	tests := []struct {
		input interface{}
		kcode byte
		buf   []byte
		err   error
	}{
		{nil, 			0x00, []byte{}, nil},
		{4, 			0x10, SBytes("04"), nil},
		{7777777, 		0x10, SBytes("f1 db da 03"), nil},
		{"foo", 		0x20, SBytes("03 66 6f 6f"), nil},
		{"Виагра", 		0x20, SBytes("0c d0 92 d0 b8 d0 b0 d0 b3 d1 80 d0 b0"), nil},
		{[]byte("foo"), 0x30, SBytes("03 66 6f 6f"), nil},
		{-4, 			0, []byte{}, fmt.Errorf("negative int keys are not supported")},
		{true,			0, []byte{}, fmt.Errorf("unknown key type (not nil/int/str/bytes)")},
	}
	for _, test := range tests {
		kcode, buf, err := EncodeKey(test.input)
		assert.Equal(t, test.kcode, kcode)
		assert.Equal(t, test.buf, buf)
		assert.Equal(t, test.err, err)
	}
}















// =====================================================================================================================
// = Two different kinds of building byte buffers.
// = They seem to be about the same performance based on the benchmarks down below.

// go test -run AppendVariadic

func TestAppendVariadic(t *testing.T) {
	z := []byte("\x01\x02\x03\x04\x05\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f")
	y := []byte("\x11\x12\x13\x14\x15\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f")
	m := byte(30)
	n := byte(60)
	var o []byte = nil             // nil slice
	var p []byte = make([]byte, 0) // empty slice

	a := make([]byte, 0)
	a = append(a, m)
	a = append(a, n)
	a = append(a, o...)
	a = append(a, p...)
	a = append(a, z...) // variadic args thing for appending a byte slice to a byte slice
	a = append(a, y...)
	// a = append(a, c, d)  // we can append multiple bytes
	// The variadic args thing is performant and reference-y, its NOT actually unpacking the bytes like xargs would.
	// fmt.Println(a)
}

func TestAppendBuffer(t *testing.T) {
	z := []byte("\x01\x02\x03\x04\x05\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f")
	y := []byte("\x11\x12\x13\x14\x15\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f")
	m := byte(30)
	n := byte(60)
	var o []byte = nil             // nil slice
	var p []byte = make([]byte, 0) // empty slice

	// from https: // golang.org/src/bytes/example_test.go
	var a bytes.Buffer // "A Buffer needs no initialization."
	a.WriteByte(m)
	a.WriteByte(n)
	a.Write(o)
	a.Write(p)
	a.Write(z)
	a.Write(y)
	fmt.Println(a.Bytes())
}

// The go wiki says:
// the nil slice is the preferred style.
// Note that there are limited circumstances where a non-nil but zero-length slice is preferred, such as when encoding JSON objects (a nil slice encodes to null, while []string{} encodes to the JSON array []).
// When designing interfaces, avoid making a distinction between a nil slice and a non-nil, zero-length slice, as this can lead to subtle programming errors.
// The Go wiki:  https://github.com/golang/go/wiki/CodeReviewComments#declaring-empty-slices

// =====================================================================================================================
// = Some Benchmarking

// go test -bench AppendVariadic
// -run and -bench are regexes. its ok to use .

// These stop getting picked up if you put them in a file named e.g. item_header_bench.go, sigh.

// Go garbage collection:
// GOCG= is a "target %age, collection is triggered when ratio of freshly allocated data to data from last time >= this percentage"
// set GOCG=off to disable, but then benchmarks OOM out.
// the default is apparently 100, which seems to be swamping these benchmarks here with GC activity.
// if i set GOGC=10000 these benchmarks get nice and deterministic without OOMing out.

func BenchmarkAppendVariadic(b *testing.B) {
	z := []byte("\x01\x02\x03\x04\x05\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f")
	y := []byte("\x11\x12\x13\x14\x15\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f")
	m := byte(30)
	n := byte(60)
	// fmt.Println("called with N = ",b.N) 		// Go's N: 100, 10_000, 1_000_000, 7_481_361

	for i := 0; i < b.N; i++ {
		a := make([]byte, 0)
		a = append(a, m)
		a = append(a, n)
		a = append(a, z...) // variadic args thing for appending a byte slice to a byte slice
		a = append(a, y...)
		_ = a // This is how we stop "variable declared and not used" errors.
	}
}

func BenchmarkAppendBuffer(b *testing.B) {
	z := []byte("\x01\x02\x03\x04\x05\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f")
	y := []byte("\x11\x12\x13\x14\x15\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f")
	m := byte(30)
	n := byte(60)

	for i := 0; i < b.N; i++ {
		var a bytes.Buffer // "A Buffer needs no initialization."
		a.WriteByte(m)
		a.WriteByte(n)
		a.Write(z)
		a.Write(y)
		_ = a.Bytes()
	}
}

