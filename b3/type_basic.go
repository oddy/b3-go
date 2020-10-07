package b3

import (
	"encoding/binary"
	"math"

	"github.com/pkg/errors"
)

// ===================== Temporary B3 basic decoders ===========================

// in go, strings are already utf8 []bytes really.
// go only utf8-decodes in 2 places 1) for i,r := range s (yielding runes), 2) casting []rune(s).
// In those instances invalide utf8 is replaces with U+FFFD (65533 utf8.RuneError) and the ops *do not crash*.
// https://stackoverflow.com/questions/34861479/how-to-detect-when-bytes-cant-be-converted-to-string-in-go

func DecodeUtf8(buf []byte) (interface{}, error) {
	return string(buf),nil
}
func DecodeBytes(buf []byte) (interface{}, error) {
	return buf,nil											// a no-op but interface{} is returned.
}

func CodecDecodeUvarint(buf []byte) (interface{}, error) {
	n, _, err := DecodeUvarint(buf)						// we dont need bytesUsed because we're sized already.S
	return n,err
}

type B3DecodeFunc func([]byte) (interface{}, error)
type B3EncodeFunc func(interface{}) ([]byte, error)

const B3_BYTES	 = 3
const B3_UTF8	 = 4
const B3_UVARINT = 7

var B3_DECODE_FUNCS = map[int]B3DecodeFunc{
	B3_BYTES:	DecodeBytes,
	B3_UTF8:	DecodeUtf8,
	B3_UVARINT:	CodecDecodeUvarint,
}

var B3_ENCODE_FUNCS = map[int]B3EncodeFunc{
	B3_BYTES:	EncodeBytes,
	B3_UTF8:	EncodeUtf8,
	B3_UVARINT:	CodecEncodeUvarint,
}

var B3_TYPE_NAMES_TO_NUMBERS = map[string]int {
	"BYTES": 3,
	"UTF8": 4,
	"UVARINT":7,
}

// ===================== Temporary B3 basic decoders ===========================




// Method: Encoders assemble [][]byte of []byte, then bytes.Join() them. We take advantage of this often for empty/nonexistant fields etc.
// Method: Decoders always take the a slice, and do NOT have to return an updated index.
// Note:   opposite to python, slicing is cheap, so we API using slices instead of a buffer+index.
//         In this respect, the Go function apis/contract/signatures are cleaner.

// Policy: Encoders MAY return no bytes to signify a Compact Zero Value (optional)
// Policu: Decoders MUST accept len(buf)==0 and return a Zero value (mandatory)
// Policy: Favouring simplicity over performance by having the type safety checks here.


// up-level wants interface{} to come in to the encoders.
// we type-assertion them and have to return an error if the type conversion doesn't pan out.

// otherwise up-level has to precheck. Someone has to concretize our inputs, its either us
// or up-level.

// Point is all these functions have to have the same signature. S

// The encoder type-pipeline from structs looks like:
// Struct  -> Reflect.Value (field) -> interface{} -> concrete type -> byte encoded buffer
// the pipeline from map[key]interface{} looks like:
// map                              -> interface{} -> concrete type -> byte encoded buffer

// with the map we would have to inspect the value and use it's Kind() or something to select the
// appropriate encoder.
// with the struct we are using b3 type-name struct tags to drive selection of encoders.

func CodecEncodeUvarint(ifValue interface{}) ([]byte, error) {
	value,ok := ifValue.(int)
	if !ok {
		return nil, errors.New("EncodeUvarint input not int")
	}
	out := EncodeUvarint(value)
	return out,nil
}


func EncodeBytes(ifValue interface{}) ([]byte, error) {
	value,ok := ifValue.([]byte)
	if !ok {
		return nil, errors.New("EncodeBytes input not []byte")
	}
	return value,nil		// direct pass-through, pretty much
}


func EncodeBool(ifValue interface{}) ([]byte, error) {
	value,ok := ifValue.(bool)
	if !ok {
		return nil, errors.New("EncodeBool input not bool")
	}
	if value {
		return []byte{0x01}, nil
	} else {
		return []byte{}, nil						// Compact zero-value for false.
	}
}


func EncodeUtf8(ifValue interface{}) ([]byte, error) {
	value,ok := ifValue.(string)
	if !ok {
		return nil, errors.New("EncodeUtf8 input not string")
	}
	return []byte(value), nil					// Strings in go are already utf8 byte arrays, score!
	// note that this is a copy
}

// "Converting between int64 and uint64 doesn't change the sign bit, only the way it's interpreted."
// This seems to work! It matched the python bytes out anyway.

func EncodeInt64(ifValue interface{}) ([]byte, error) {
	value,ok := ifValue.(uint64)
	if !ok {
		return nil, errors.New("EncodeInt64 input not convertable to int64")
	}
	if value == 0 {
		return []byte{}, nil											// output compact zero value
	}
	out := make([]byte, 8)
	binary.LittleEndian.PutUint64(out, value)
	return out,nil
}

func EncodeFloat64(ifValue interface{}) ([]byte, error) {
	value,ok := ifValue.(float64)
	if !ok {
		return nil, errors.New("EncodeFloat64 input not convertable to float64")
	}

	if value == 0 {
		return []byte{}, nil									// CZV
	}
	outU64 := math.Float64bits(value)
	out := make([]byte, 8)
	binary.LittleEndian.PutUint64(out, outU64)
	return out, nil
}


// Currently Stamp64 only accepts time.Time. If we wanted it to accept other things,
// it would grow an error return which means all the other encoders would grow an error return too.
// this may still happen.

// come back to this one.
//
/*
func EncodeStamp64(value time.Time) []byte {
	nano := value.UnixNano()
	out := make([]byte, 8)
	binary.LittleEndian.PutUint64(out, uint64(nano))
	return out
}
*/

func EncodeComplex(ifValue interface{}) ([]byte, error) {
	value,ok := ifValue.(complex128)
	if !ok {
		return nil, errors.New("EncodeComplex input not convertable to complex128")
	}

	if value == 0 {								// confirmed this works, nice syntactic sugar
		return []byte{}, nil
	}
	out := make([]byte, 16)
	binary.LittleEndian.PutUint64(out,     math.Float64bits(real(value)))
	binary.LittleEndian.PutUint64(out[8:], math.Float64bits(imag(value)))
	return out, nil
}



