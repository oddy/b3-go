package b3

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/pkg/errors"
)

// fields with no b3 struct tags are ignored.
// fields not present in the incoming data are ignored (will be 0 or whatever the incoming struct already has)


func BufToStruct(buf []byte, dataLen int, destStructPtr interface{}) error {
	var ok bool

	// Get the struct pointer from the interface{}
	ptr := reflect.ValueOf(destStructPtr)
	// must be a pointer, if we call Elem on non-pointer, Elem panics
	if ptr.Kind() != reflect.Ptr {
		return errors.New("destStructPtr must be a pointer")
	}

	// must be a struct, NumField panics if called on a non-struct
	destStruct := ptr.Elem()
	if destStruct.Kind() != reflect.Struct {
		return errors.New("destStructPtr must be a pointer to a struct")
	}

	// we need this to get at the b3 struct tags.
	destStructType := reflect.TypeOf(destStructPtr).Elem()

	index := 0
	for index < len(buf) {
		hdr, bytesUsed, err := DecodeHeader(buf[index:])
		if err != nil {
			return errors.Wrap(err, "fillstruct decode header fail")
		}
		index += bytesUsed
		// fmt.Println("filllstruct got header ",hdr)
		// [hdr]   DataType, Key(tag), IsNull, DataLen

		// Policy:  key type must be int.
		// Todo:    support for string and maybe bytes key types.
		tag,kok := hdr.Key.(int)
		if !kok {
			return errors.New("only int keys supported")
		}

		// use data type to get b3 decoder.
		DecodeFunc,fok := B3_DECODE_FUNCS[hdr.DataType]
		if !fok {
			return errors.New("no decoder found for data type")
		}

		// b3 decode item data to interface value.
		// Policy:  incoming b3 nulls -> go zero-values.
		//          otherwise "cannot use nil as type int in field value"
		var decodedValue interface{}
		if hdr.IsNull {
			decodedValue,err = DecodeFunc([]byte{})		// []byte{} = empty slice,  []byte = nil slice. we want empty not nil.
		} else {										// i dont think we can pass []byte by itself anyway, wont compile
			decodedValue,err = DecodeFunc(buf[index:index+hdr.DataLen])
		}
		if err != nil {
			return errors.Wrap(err, "b3 type decoder fail")
		}
		index += hdr.DataLen

		// fmt.Println("key/tag number ",tag)
		// fmt.Println("decoded value  ",decodedValue)

		// with the struct we're given, find the field using struct tags b3.tag

		// Search struct for the matching field.
		fieldFound := false
		fieldNum := 0
		fieldB3TypeInt := 0			// not a valid type.
		for ; fieldNum < destStruct.NumField() ; fieldNum++ {

			// Get struct tags b3.tag 'number'
			tfield := destStructType.Field(fieldNum)
			fieldB3Tag := tfield.Tag.Get("b3.tag")
			if fieldB3Tag == "" {
				continue								// no b3.tag struct tag, skip struct field.
			}
			fieldB3TagNum,fberr := strconv.Atoi(fieldB3Tag)
			if fberr != nil {
				return errors.Wrap(fberr, "struct b3.tag is not a number")
				//continue								// cant convert struct tag to int, skip struct field (?)
			}
			if fieldB3TagNum == tag {		// found it!

				// extract the b3.type struct tag too.
				fieldB3Type := tfield.Tag.Get("b3.type")
				if fieldB3Type == "" {
					return errors.New("struct b3.type is missing")
				}
				fieldB3TypeInt, ok = B3_TYPE_NAMES_TO_NUMBERS[fieldB3Type]
				if !ok {
					return errors.New("struct b3.type name not found in b3 types")
				}
				fieldFound = true
				break
			}
		}
		if !fieldFound {	// wanted b3 tag not found in struct, ignore
			fmt.Println("b3 tag not found in struct tags, ignoring ",hdr.Key)
			continue
		}

		// fieldNum now has the number of the struct field.
		// ensure the field is valid and settable.
		fieldVal := destStruct.Field(fieldNum)
		if !fieldVal.IsValid() {
			return errors.New("struct field is not valid")
		}
		if !fieldVal.CanSet() {
			return errors.New("struct field is not settable")
		}

		// ensure the b3 types match!
		if hdr.DataType != fieldB3TypeInt {
			return errors.New("struct field b3 type mismatch vs incoming data type")
		}

		// ---- Actually set it, woo! ----
		refVal := reflect.ValueOf(decodedValue)
		fieldVal.Set(refVal)

		fmt.Println("struct field number ",fieldNum," name ",destStructType.Field(fieldNum).Name, " successfully set val to ",decodedValue)

	}
	return nil
}



func StructToBuf(srcStructIf interface{}) ([]byte, error) {
	// ensure srcStruct is actually a struct
	srcStruct := reflect.ValueOf(srcStructIf)
	if srcStruct.Kind() != reflect.Struct {
		return nil,errors.New("input must be a struct")
	}
	// we need this to get at the b3 struct tags.
	srcStructType := reflect.TypeOf(srcStructIf)

	fmt.Println("ok got struct")
	fmt.Println(srcStruct)
	fmt.Println(srcStructType)

	// go through the struct fields, encode the values and keys, make a bunch of item buffers
	// put the item buffers into a map

	itemHdrBufs := make(map[int][]byte)			// keyed by b3 tag number
	itemValBufs := make(map[int][]byte)			// keyed by b3 tag number
	itemKeys := make([]int, 0, 10)				// to be sorted

	for fieldNum := 0 ; fieldNum < srcStruct.NumField() ; fieldNum++ {
		// Get struct tags b3.tag 'number'
		tfield := srcStructType.Field(fieldNum)
		fieldB3Tag := tfield.Tag.Get("b3.tag")
		if fieldB3Tag == "" {
			continue								// no b3.tag struct tag, skip struct field.
		}
		// turn into actual number
		fieldB3TagNum,fberr := strconv.Atoi(fieldB3Tag)
		if fberr != nil {
			return nil, errors.Wrap(fberr, "struct b3.tag is not a number")
		}
		// get b3.type name
		fieldB3TypeName := tfield.Tag.Get("b3.type")
		if fieldB3TypeName == "" {
			return nil, errors.New("struct b3.type is invalid")
		}
		// turn into type number
		fieldB3TypeInt, ok := B3_TYPE_NAMES_TO_NUMBERS[fieldB3TypeName]
		if !ok {
			return nil, errors.New("struct b3.type name not found in b3 types")
		}

		// so fieldB3TagNum is the key
		// now encode the value

		// we get the value from the struct as a reflect.Value
		fieldVal := srcStruct.Field(fieldNum)
		fmt.Println(" field ",fieldNum," val ",fieldVal)
		fmt.Printf(" field val   is a %T\n", fieldVal)

		// Turn the value into an interface value for feeding to the decoders
		fieldIfVal := fieldVal.Interface()	// The encoder functions take interface{} and type check themselves.
		fmt.Printf(" field ifVal is a %T\n", fieldIfVal)

		// Select encoder based on struct tag b3.type number.
		// (The encoder funcs then type-assert the value to ensure it's the right concrete type.)

		// use data type to get b3 encoder.
		EncodeFunc,fok := B3_ENCODE_FUNCS[fieldB3TypeInt]
		if !fok {
			return nil,errors.New("no encoder found for b3.type")
		}

		// Feed the value to the b3 decoders
		valBuf,err := EncodeFunc(fieldIfVal)
		if err != nil {
			return nil, errors.Wrap(err, "data value encode fail")
		}

		// Make b3 item header for value
		itmHdr := ItemHeader{DataType: fieldB3TypeInt, Key: fieldB3TagNum, IsNull: false, DataLen: len(valBuf)}

		hdrBuf,herr := EncodeHeader(itmHdr)
		if herr != nil {
			return nil, errors.Wrap(err, "b3 item header encode fail")
		}

		// Stash item hdr & value bytes in map by key/tag number
		itemHdrBufs[fieldB3TagNum] = hdrBuf
		itemValBufs[fieldB3TagNum] = valBuf
		// Stash the key numbers in a slice so we can sort them
		itemKeys = append(itemKeys, fieldB3TagNum)

	}

	if len(itemValBufs) == 0 {
		return nil, errors.New("no struct fields were successfully encoded")
	}

	fmt.Println("item header bufs ",itemHdrBufs)
	fmt.Println("item val bufs    ",itemValBufs)

	// sort the itemBufs keys
	fmt.Println("keys before sort ",itemKeys)
	sort.Ints(itemKeys)
	fmt.Println("keys after sort  ",itemKeys)

	// Then range through the keys in sorted order and just append the itemBufs into a superbuf and return that
	outBuf := make([]byte,0) //, 0, 64)			// try and keep it on the stack for small messages (?)
	for _,kn := range itemKeys {
		fmt.Println("kn ",kn)
		outBuf = append(outBuf, itemHdrBufs[kn]...)
		fmt.Print(Hexdump(outBuf, len(outBuf)))
		fmt.Println()
		outBuf = append(outBuf, itemValBufs[kn]...)
		fmt.Print(Hexdump(outBuf, len(outBuf)))
		fmt.Println()
	}

	fmt.Println("Final output buf, len = ",len(outBuf))
	fmt.Print(Hexdump(outBuf, len(outBuf)))
	fmt.Println()

	return outBuf,nil
}
