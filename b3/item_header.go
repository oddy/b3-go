package b3

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "go/types"

	"github.com/pkg/errors"
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
 | is null    | has data   |9
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

// Policy: DONT GET FANCY
// Policy: returning byteslices everywhere, then return bytes.Join( [][]byte{x,y,z} , nil )
//         Because thats how the python code does it and it will make the code very simple and straight-port.
//         See "Journey of pain" below for how we got here.

// Policy: Screw it, use int everywhere we can. "int" is the default type and it will mean a lot less casting.

type ItemHeader struct {
	DataType int
	Key interface{}
	IsNull bool
	DataLen int		// 0 = not hasData on encode side, len forced 0 if hasData FALSE on decode side.
}

func EncodeHeader(hdr ItemHeader) ([]byte, error) {
	var extDataTypeBytes []byte
	var lenBytes []byte
	var cbyte byte

	// --- Null & data len ---
	if hdr.IsNull {
		cbyte |= 0x80 								// data value is null. Note: null supercedes has-data
	} else {
		if hdr.DataLen > 0 {
			cbyte |= 0x40							// has data flag on
			lenBytes = append(lenBytes, EncodeUvarint(hdr.DataLen)...)
		}
	}
	// fmt.Printf("cbyte is %02x\n",cbyte)

	// --- Key type ---
	keyTypeBits, keyBytes, err := EncodeKey(hdr.Key)
	if err != nil {
		return []byte{}, err
	}
	cbyte |= keyTypeBits & 0x30					// middle 2 bits for key type

	// fmt.Printf("cbyte is %02x\n",cbyte)

	// --- Data type ---
	if hdr.DataType < 0 {							// Sanity S
		return []byte{}, fmt.Errorf("-ve data types not permitted")
	}

	if hdr.DataType > 14 { 							// 'extended' data types 15 and up are a seperate uvarint
		extDataTypeBytes = append(extDataTypeBytes, EncodeUvarint(hdr.DataType)...)
		cbyte |= 0x0f 							// control byte data_typeck bits set to all 1's to signify this
	} else {
		cbyte |= byte(hdr.DataType) & 0x0f
	}

	// --- Build header ---
	// fmt.Printf("cbyte is %02x\n",cbyte)
	out := bytes.Join([][]byte{ []byte{cbyte}, extDataTypeBytes, keyBytes, lenBytes  }, nil)
	return out, nil

}



// todo: its still slightly up in the air what index we return if there is an error.
//       in python, decode_header exceptions are unhandled even by the composite unpackers, so it blows straight
//       through to user code. So there's no actual answer yet, but going forward we should maintain a policy of:
// policy: "all returns are invalid if err != nil"



// Gonna do this with an interface and a typeswitch.
// "The zero value of a slice is nil". Also there are "nil slices" and "empty slices".
// You can cast a -ve into to a uint, you get a yuuge number. So it lets you do it and "C's you up"

func EncodeKey(ikey interface{}) (byte, []byte, error) {

	_ = json.Encoder()

	switch key := ikey.(type) {



	case nil: // also nil slice and/or empty slice?		// does this work?
		return 0x00, []byte{}, nil


	// note:   if you e.g. "case int,uint:"  go doesn't concretize and you get interface{}
	// policy: only accepting ints for now, prefer Simplicity over flexibility(?)
	case int:
		if key < 0 {
			return 0, []byte{}, fmt.Errorf("negative int keys are not supported")
		}
		return 0x10, EncodeUvarint(key), nil

	case string:
		keyBytes := []byte(key) // like strings ARE utf8 bytes sooo this should be ok
		lenBytes := EncodeUvarint(len(keyBytes))
		return 0x20, bytes.Join([][]byte{lenBytes, keyBytes}, nil), nil

	case []byte:
		lenBytes := EncodeUvarint(len(key))
		return 0x30, bytes.Join([][]byte{lenBytes, key}, nil), nil

	default:
		return 0, []byte{}, fmt.Errorf("unknown key type (not nil/int/str/bytes)")
	}
}

// we have to precheck slices, because exceeding limits is a panic!, instead of a besteffort like in python.

// nextIndex should be pointing to the location of the start of the key, so keystuff[0]/
// if there is no keystuff, then nextIndex will == len(buf)

// Go is a little weird - in a len(3) buf e.g. "foo",
// [3] blows up as you'd expect (next char after final 'o'), [3:4] blows up too (4 > len), BUT
// [3:3] is ok and returns [].

// Do we do a lot of error checking, or do we make a slice-function that acts like python's does?


// the DEcoders are going to just be given slices. The bounds-checking will be done by the codec's caller.


// "it’s idiomatic to have functions like slice = doSomethingWithSlice(slice) and less so to see doSomethingWithSlice(&slice)"


// We don't need to pass buf and index if we're passing slices around all the time. Just pass a new slice.
// You can see in DecodeUvarint










// ++++ new +++++++

	// ==============================================================================================
	// We're passing slices. decode_header is special, it gets the [x:] rest-of-buf,
	// everything else gets [x:y] because sizes are KNOWN for everything else.
	// ==============================================================================================

// decode_header DOES need to return number of bytes consumed, but the size-known functions dont.
// DECIDED.

// Q: Do errors return 0 bytes consumed?
// A: yes. bytesConsumed is invalid if there is an error. Return 0 for it and expect it not to be used.

func DecodeKey(keyTypeBits byte, buf []byte) (interface{}, int, error) {  // Return: key-value, bytes-consumed, error
	if keyTypeBits == 0x00 {							// no key
		return nil, 0, nil
	}

	if keyTypeBits == 0x10 {							// (u)int key
		return DecodeUvarint(buf)		// Note also would return error
	}

	if keyTypeBits == 0x20 || keyTypeBits == 0x30 {		// string or bytes key.
		klen, nLenBytes, err := DecodeUvarint(buf)		// nLenBytes = how many bytes the uvarint len itself is.
		if err != nil {
			return nil, 0, errors.Wrap(err,"decodekey decode len uvarint")  // bytesConsumed should be 0 if error.
		}

		// result returned from DecodeUvarint will never be negative.

		end := nLenBytes + klen
		if end >= len(buf) {
			return nil, 0, errors.New("key size > buffer len")
		}

		keyBytes := buf[nLenBytes : end]

		if keyTypeBits == 0x30 {
			return keyBytes, end, nil
		} else {
			return string(keyBytes), end, nil
		}

	}

	return nil, 0, errors.New("invalid key type in control byte")
}


// 							   returns: ItemHeader struct, bytesUsed int, error

func DecodeHeader(buf []byte) (ItemHeader, int, error) {
	var index, bytesUsed int
	var err error
	var key interface{}

	hdr := ItemHeader{}
	// Must be at least 1 byte
	if len(buf) < 1 {
		return hdr,0,errors.New("decodeheader buf empty")
	}
	cbyte := buf[0]				// control byte
	index += 1

	// --- data type ---
	hdr.DataType = int(cbyte & 0x0f)
	if hdr.DataType == 15 {
		hdr.DataType, bytesUsed, err = DecodeUvarint(buf[index:])
		if err != nil {
			return hdr,0,errors.Wrap(err,"item header extended datatype decode failed")
		}
		index += bytesUsed
	}

	// --- Key ---
	keyTypeBits := cbyte & 0x30
	hdr.Key, bytesUsed, err = DecodeKey(keyTypeBits, buf[index:])
	if err != nil {
		return hdr,0,errors.Wrap(err,"item header decode key fail")
	}
	index += bytesUsed

	// --- Null check ---
	hdr.IsNull  = (cbyte & 0x80) == 0x80
	hasData    := (cbyte & 0x40) == 0x40
	if hdr.IsNull && hasData {
		return hdr,0,errors.New("item header invalid state - is_null and has_data both ON")
	}

	// --- Data len ---
	hdr.DataLen = 0
	if hasData {
		hdr.DataLen, bytesUsed, err = DecodeUvarint(buf[index:])
		if err != nil {
			return hdr,0,errors.Wrap(err, "item header decode data len fail")
		}
		index += bytesUsed
	}

	return hdr,index,nil
}


// Remember reallocations are also copies.

// ========= Journey of pain (delete this later) - aka how are we building buffers ====================

// The whole thing is either return byte buffers and glue them at the end, like python.
// or Write to things. Looks like the Go Way is 'Write to things'. Start with a bytes.Buffer up top, pass it everywhere

// in python everything is a reference, in go very much not.
// SO we either do lots of call and return, or we pass references using pointers.

// IF we used []byte we'd HAVE to do the old    a = append(a) and return a everywhere.
// we can't use pointers with []bytes because append doesn't roll pointers. the value that comes back is different.

// IF we use *bytes.Buffer, we can probably just .Write, and save ourselves a lot of lines of code.

// --OR NOT--

// *****************************************************************************************
// in encode_header, if you want to build the buffer, you have to PREpend cbyte, which means a copy right at the end, to bump everything right-one
// because go isn't so hot at prepends.
// We could get cunning and leave a byte free at the start and drop it in later but thats not the point right now.

// And if we do it that way we HAVE to do a lot of CODE in the Right Order. because then !! code order is buffer order !!.
// so we have to contort code order so we can use *bytes.Buffer and Write things in the right order.
// *** I'm not really down with this, too much thinking. ***

// vs
// return-byte-slices,  where we make an array of them then call Join and
// THE SAME AMOUNT OF COPYING HAPPENS ANYWAY.  (at the Join).

// ++++++++++++++++++++++++++++++++++++++++++++++++++++
// if we return byte slices and join() them at the end of functions to make a single byte slice, then
// WE CAN DIRECT PORT THE PYTHON CODE, and not THINK
// Because thats what how the python code works.
// ++++++++++++++++++++++++++++++++++++++++++++++++++++

// z := [][]byte{a, b}
// fmt.Println(z)
// y := bytes.Join(z, []byte{0x00})
// fmt.Println(y)
// x := bytes.Join(z, nil)
// fmt.Println(x)

// So we're gonna Not Get Fancy and just have a lot of byte slice returning going on and concatenating them etc when we need to.
// SO item header and the little codecs at least, Just Return A Byte Slice.
