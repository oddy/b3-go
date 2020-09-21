package b3

import (
	"bytes"
	"fmt"
	_ "go/types"
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


func EncodeHeader(dataType int, key interface{}, isNull bool, dataLen int) ([]byte, error) {
	var extDataTypeBytes []byte
	var lenBytes []byte
	var cbyte byte

	// --- Null & data len ---
	if isNull {
		cbyte |= 0x80 								// data value is null. Note: null supercedes has-data
	} else {
		if dataLen > 0 {
			cbyte |= 0x40							// has data flag on
			lenBytes = append(lenBytes, EncodeUvarint(dataLen)...)
		}
	}
	// fmt.Printf("cbyte is %02x\n",cbyte)


	// --- Key type ---
	keyTypeBits, keyBytes, err := EncodeKey(key)
	if err != nil {
		return []byte{}, err
	}
	cbyte |= keyTypeBits & 0x30					// middle 2 bits for key type

	// fmt.Printf("cbyte is %02x\n",cbyte)

	// --- Data type ---
	if dataType < 0 {							// Sanity S
		return []byte{}, fmt.Errorf("-ve data types not permitted")
	}

	if dataType > 14 { 							// 'extended' data types 15 and up are a seperate uvarint
		extDataTypeBytes = append(extDataTypeBytes, EncodeUvarint(dataType)...)
		cbyte |= 0x0f 							// control byte data_typeck bits set to all 1's to signify this
	} else {
		cbyte |= byte(dataType) & 0x0f
	}

	// --- Build header ---
	// fmt.Printf("cbyte is %02x\n",cbyte)
	out := bytes.Join([][]byte{ []byte{cbyte}, extDataTypeBytes, keyBytes, lenBytes  }, nil)
	return out, nil

}




// Gonna do this with an interface and a typeswitch.
// "The zero value of a slice is nil". Also there are "nil slices" and "empty slices".
// You can cast a -ve into to a uint, you get a yuuge number. So it lets you do it and "C's you up"

func EncodeKey(ikey interface{}) (byte, []byte, error) {
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


func DecodeKey(keyTypeBits byte, buf []byte, index int) (interface{}, int, error) {
	if keyTypeBits == 0x00 {
		return nil, index, nil
	}

	if keyTypeBits == 0x10 {
		return DecodeUvarint(buf, index)					// Note also would return error
	}

	if keyTypeBits == 0x20 {
		klen, nextIndex, err := DecodeUvarint(buf, index)
		if err != nil {
			return nil, index, err
		}
		keyStr := string(buf[nextIndex : nextIndex + klen])
		return keyStr, nextIndex+klen, nil
	}

	if keyTypeBits == 0x30 {
		klen, nextIndex, err := DecodeUvarint(buf, index)
		if err != nil {
			return nil, index, err
		}
		keyBytes := buf[nextIndex : nextIndex + klen]
		return keyBytes, nextIndex+klen, nil
	}

	return nil, index, fmt.Errorf("invalid key type in control byte")
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
