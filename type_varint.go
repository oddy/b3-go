package b3

import (
	"fmt"
)


// Policy: using uint64 for all numbers. (max interop on 32bit even tho slower).
//         possibly revisit this. If 32bit performance ends up being a thing (if 32bit ends up being a thing).

// ===== Encoding =========

// Policy: Not enough buffer isn't an error because we're append() ing

func EncodeUvarint(x int) []byte  {				// todo: figure out about uints vs ints coming in here.
	var out []byte
	for x >= 0x80 {
		out = append(out, byte(x)|0x80)
		x >>= 7
	}
	out = append(out,byte(x))
	return out
}

func EncodeSvarint(x int)  []byte {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	return EncodeUvarint(int(ux))
}


// ========= Decoding ==========


// Policy: indexes are ints now because thats what for:=range shits out.

// --- Decoding into fixed-size numeric variables and pre-checking to stave off overflow panics. ---
// (2^64)-1  =  18_446_744_073_709_551_615  =  \xff\xff\xff\xff\xff\xff\xff\xff\xff\x01
// (2^64)    =  18_446_744_073_709_551_616  =  \x80\x80\x80\x80\x80\x80\x80\x80\x80\x02
// Hence the if i > 9 || i == 9 && byt > 1 {  return 0,0, fmt.Errorf("uvarint > uint64") }

// For "over int" its easier because...
// (2^63)-1  =   9_223_372_036_854_775_807  =  \xff\xff\xff\xff\xff\xff\xff\xff\x7f
// (2^63)    =   9_223_372_036_854_775_808  =  \x80\x80\x80\x80\x80\x80\x80\x80\x80\x01
// ... its when i goes from 8 to 9.


func DecodeUvarint(buf []byte, index int) (int, int, error) { // returns output,index,error
	var result int
	var shift int
	buf2 := buf[index:]
	for i, byt := range buf2 {
		if byt < 0x80 { // MSbit clear, final byte.
			if i >= 9 {
				return 0, 0, fmt.Errorf("uvarint > int64")
			}
			return result | int(byt)<<shift, index + i + 1, nil // Ok
		}
		result |= int(byt&0x7f) << shift
		shift += 7
	}
	return 0, 0, fmt.Errorf("uvarint > buffer")
}

func DecodeSvarint(buf []byte, index int) (int, int, error) { // returns output,index,error
	ux, resIndex, err := DecodeUvarint(buf, index)
	if err != nil {
		return 0, 0, err
	}

	result := int(ux >> 1)
	if ux&1 != 0 {
		result = ^result
	}
	return result, resIndex, nil
}









// ==== this might be old/invalid now? ====

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


// Because not panicing means we can do stuff like have the highest level do things like disconnect the socket.


// The varints are self-sizing for the item header, so we DO have to do the buf,index thing.
// but for the CODECS, we can go the "simple buf" way.
