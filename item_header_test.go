package b3

import (
	"bytes"
	"fmt"
	"testing"
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
// = Item header keys

func TestKeytypeNoneIntEnc(t *testing.T) {

}

func TestKeytypeNoneIntDec(t *testing.T) {

}

func TestKeytypeStrBytesEnc(t *testing.T) {

}

func TestKeytypeStrBytesDec(t *testing.T) {

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
	a = append(a, z...)   // variadic args thing for appending a byte slice to a byte slice
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
		_ = a				// This is how we stop "variable declared and not used" errors.
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

