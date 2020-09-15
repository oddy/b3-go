package b3

import (
	"fmt"
)


// Policy: using uint64 for all numbers. (max interop on 32bit even tho slower).
//         possibly revisit this. If 32bit performance ends up being a thing (if 32bit ends up being a thing).

// ===== Encoding =========

// For ENcoding we take a given []byte and append() to it then return it.
// So its fast because pointer and go can do its realloc trick only when needed. Nice!

// This is really simple for indifr the functions, integrates really nicely with
// "upstairs" callers, and is still quite performant.

// It's actually a lot nicer than b3-py's "return buf fragments and join() them a lot" model,
// and is enabled by go's append() and byte-slice semantics.

// Policy: Not enough buffer isn't an error because we're append() ing

func EncodeUvarint(x uint64) []byte  {
	var out []byte
	for x >= 0x80 {
		out = append(out, byte(x)|0x80)
		x >>= 7
	}
	out = append(out,byte(x))
	return out
}

func EncodeSvarint(x int64)  []byte {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	return EncodeUvarint(ux)
}

// in Go, slicing is the low-cost activity, vs slicing being the high cost activity in python
// So maybe we pass different slices around everywhere, instead of the same slice and a pointer number?

// ========= Decoding ==========

// In python we use buf,index and return value,index
// And we do a lot of IntByteAt and

// Remember the top-levels have guarded-lengths. So they call the decoders like:
// value = DecoderFn(buf, index, index+data_len)

// If in GO, slicing is very high performance, we could chop-slice up THERE and save a LOT of call-complexity in the
// decoders!

// It's more conceptually clean to do the slicing WHERE YOU HAVE THE SIZE INFORMATION for sure.

// But: unpack_into is recursive - we could still slice for it.
//      the underlying array is still fine.
// Fancy people would probably use Byte Readers and stuff like that, but we are going for deadshit simple.

// Not gonna bother with readers and writers for now, just byte-slices.

// We should use uint64 for everything we can, and we will have to deal with overflow errors somehow. (panic??)

// should we just panic if the uvarint overflows a uint64 ?
// "Once your code is compiled, there is no difference between uint and uint64 (assuming 64-bit arch). Converting between the two is free"
// I think we should just panic -
// no think it through
// for our quick hack, we're targetting decode into struct.
// if the struct members are smaller than uint64 then we will have to deal with overflow errors.

// Policy: return uint64 for all numbers and deal with overflow errors.

// Because not panicing means we can do stuff like have the highest level do things like disconnect the socket.

// Policy: errors: overflow from numbers bigger than u/int64
// Policy: errors: the varint keep going off the end of the given buffer. We need to check for this.

// The varints are self-sizing for the item header, so we DO have to do the buf,index thing.
// but for the CODECS, we can go the "simple buf" way.

// Policy: indexes are ints now because thats what for:=range shits out.

func DecodeUvarintInternal(buf []byte, index int) (uint64, int, error) { // returns output,index,error
	var result uint64
	var shift uint
	buf2 := buf[index:]
	for i, byt := range buf2 {
		if byt < 0x80 { // MSbit clear, final byte.
			if i > 9 || i == 9 && byt > 1 { // varint was too big to fit in a uint64
				return 0, 0, fmt.Errorf("uvarint > uint64")
			}
			return result | uint64(byt)<<shift, index + i + 1, nil // Ok
		}
		result |= uint64(byt&0x7f) << shift
		shift += 7
	}
	return 0, 0, fmt.Errorf("uvarint > buffer")
}

func DecodeSvarintInternal(buf []byte, index int) (int64, int, error) { // returns output,index,error
	ux, resIndex, err := DecodeUvarintInternal(buf, index)
	if err != nil {
		return 0, 0, err
	}

	result := int64(ux >> 1)
	if ux&1 != 0 {
		result = ^result
	}
	return result, resIndex, nil

}
