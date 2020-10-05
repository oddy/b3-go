package b3

import (
	"encoding/binary"
	"math"
	"time"
)
// Method: Encoders assemble [][]byte of []byte, then bytes.Join() them. We take advantage of this often for empty/nonexistant fields etc.
// Method: Decoders always take the a slice, and do NOT have to return an updated index.
// Note:   opposite to python, slicing is cheap, so we API using slices instead of a buffer+index.
//         In this respect, the Go function apis/contract/signatures are cleaner.

// Policy: Encoders MAY return no bytes to signify a Compact Zero Value (optional)
// Policu: Decoders MUST accept len(buf)==0 and return a Zero value (mandatory)
// Policy: Favouring simplicity over performance by having the type safety checks here.

func EncodeBool(value bool) []byte {
	if value {
		return []byte{0x01}
	} else {
		return []byte{}						// Compact zero-value for false.
	}
}


func EncodeUtf8(value string) []byte {
	return []byte(value)					// Strings in go are already utf8 byte arrays, score!
	// note that this is a copy
}

// "Converting between int64 and uint64 doesn't change the sign bit, only the way it's interpreted."
// This seems to work! It matched the python bytes out anyway.

func EncodeInt64(value int64) []byte {
	if value == 0 {
		return []byte{}											// output compact zero value
	}
	out := make([]byte, 8)
	binary.LittleEndian.PutUint64(out, uint64(value))
	return out
}

func EncodeFloat64(value float64) []byte {
	if value == 0 {
		return []byte{}									// CZV
	}
	outU64 := math.Float64bits(value)
	out := make([]byte, 8)
	binary.LittleEndian.PutUint64(out, outU64)
	return out
}


// Currently Stamp64 only accepts time.Time. If we wanted it to accept other things,
// it would grow an error return which means all the other encoders would grow an error return too.
// this may still happen.

// come back to this one.
//

func EncodeStamp64(value time.Time) []byte {
	nano := value.UnixNano()
	out := make([]byte, 8)
	binary.LittleEndian.PutUint64(out, uint64(nano))
	return out
}


func EncodeComplex(value complex128) []byte {
	if value == 0 {								// confirmed this works, nice syntactic sugar
		return []byte{}
	}
	out := make([]byte, 16)
	binary.LittleEndian.PutUint64(out,     math.Float64bits(real(value)))
	binary.LittleEndian.PutUint64(out[8:], math.Float64bits(imag(value)))
	return out
}



