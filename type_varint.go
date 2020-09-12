package b3

// import "encoding/binary"

// Actual varint/uvarint encoding is supplied by go's encoding/binary module
// Tho its functions just panic if there isn't enough space, which is a bit shit.
// So maybe we'll pull them in here and make our own.


// A lot of B3's varints are small - 1 or 2 bytes
// So we're gonna dynamic alloc the byteslices ourselves for now.

// Encoders return the buf and maybe the len


// Actually we want to take an input []byte and append() to it a lot and then return it.
// This is really simple for indifr the functions, integrates really nicely with
// "upstairs" callers, and is still quite performant.

// It's actually a lot nicer than b3-py's "return buf fragments and join() them a lot" model,
// and is enabled by go's append() and byte-slice semantics.


func EncodeUvarint(x uint) ([]byte,uint) {
	buf := make([]byte, 1)
	i := uint(0)
	for x >= 0x80 {
		buf = append(buf, byte(x) | 0x80)
		x >>= 7
		i++
	}
	buf = append(buf, byte(x))
	return buf, i+1
}


