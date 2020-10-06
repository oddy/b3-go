package main

import (
	"fmt"
	"log"
	"math/bits"
	"net"
	"reflect"
	"time"

	"github.com/pkg/errors"
)

// Note: BMQ framing (outermost frame) vs BMQ-LL (an inner protocol for link-local messages).

// Keeping it simple. IF a REconnect fails, fatal/panic.

// Main loop
// Connect 'loop'
// Receive 'loop'. LOTS of socket receive calls. Nothing clever.
//                 Go is fast, so we can get away with receiving a byte at a time for a bit.

// I say we do the same in python too. It wont be performant, but we don't really need it to be.
// And when we do, a little C lib will actually work for managing connections and

//const TIMEOUT = 2 * time.Minute		// prod
const TIMEOUT = 2 * time.Second			// testing
const CONNECT_TIMEOUT = 15 * time.Second

func ConnectLoop() {
	for {
		fmt.Println("(re)connecting...")
		conn, cerr := net.DialTimeout("tcp", "127.0.0.1:7777", CONNECT_TIMEOUT)
		must(cerr)								// Connection fail is fatal.
		fmt.Println("Connected")

		err := CommsLoop(conn)					// returns nil if we've been told to shut down.

		_ = conn.Close()
		if err == nil {
			fmt.Println("Shutdown was requested, finishing")
			break
		}
		fmt.Println("Comms error: ",err)
	}
}

func ReceiveByte(conn net.Conn) (byte, error) {
	var err error
	if err = conn.SetReadDeadline(time.Now().Add(TIMEOUT)); err != nil {
		return 0x00, errors.Wrap(err, "set timout")
	}
	wbuf := make([]byte, 1)
	_,err = conn.Read(wbuf)
	if err != nil {				// Note this includes timeouts
		return 0x00, errors.Wrap(err, "read")
	}
	return wbuf[0], nil
}

func Expect(conn net.Conn, wants []byte) (byte, error) {
	b, err := ReceiveByte(conn)
	if err != nil {
		return 0x00, errors.Wrap(err, "expect")
	}
	for _,want := range wants {
		if b == want {
			return b, nil // Success!
		}
	}
	return 0x00, errors.New("incorrect byte")
}

// +------------+------------+------------+------------+------------+------------+------------+------------+
// | is null    | has data   | key type   | key type   | data type  | data type  | data type  | data type  |
// +------------+------------+------------+------------+------------+------------+------------+------------+

// Policy: Framing uses a B3 item header, header's datalen is the length of the payload in bytes.
//         header data type currently locked on DICT
//         header Key signals which inner-protocol the payload is. (Mandatory integer-only key).
//         The only inner-protocol supported is BMQ-LL (0x69).

// Policy: this means our framing header is actually always 0101 0001, 0x69, uvarint(datalen).

// ALL timeouts are errors (kill & reconnect) APART FROM the first-byte timeout where we are idle waiting for a message.
// === BUT WAIT ===
// I think we should make it extra easy on ourselves and not even send pings, just kill the connection after a long timeout.
// If we dont have to send pings either, and rely on the mothership watchdog-pinging us, we have have 1 timeout = error for
// literally everything. Make it say 2 minutes. Max time the app will stick around for if it somehow doesn't detect the
// mothership disconnecting.

// So all timeouts are 2 minutes, and all timeouts are an error (kill & reconnect).
// This means we dont have to handle timeouts at all.

// If we want a ping that is a different inner-protocol we'd have to Expect different bytes, idk if i can be bothered.
// but the pings are for us, not for the message processor. They're just to keep us from timing out.
// so use a different inner-protocol number. hex 50 is a capital P.

// key of 0x50 is ping/keepalive "inner protocol". No inner message, null flag on. data type doesn't matter, probably set BYTES

// null on, no data, no key, dont-care datatype (svarint is 88)
// 88
// a dict but it's nil? classical b3 impls will complain that the top level isn't a dict or list.



// if we're

// Literally just triggering a receive. If it's 88, we want to just reup our timeout and continue, that is all


// one byte, nil. no data, no key, data type svarint. Do it.

func ReadUvarint(conn net.Conn) (int, error) {
	var result int
	var shift int
	var i int
	var byt byte
	var err error

	for {
		// fmt.Println("readUvarint i ",i)
		byt, err = ReceiveByte(conn)
		if err != nil {
			return 0, errors.Wrap(err, "rx byte")
		}

		if byt < 0x80 { // MSbit clear, final byte.
			if i >= 9 {
				return 0, errors.New("uvarint > int64")
			}
			return result | int(byt)<<shift, nil 	// Success
		}
		result |= int(byt&0x7f) << shift
		shift += 7
		i++
	}
	// return 0, errors.New("uvarint > buffer")		// unreachable
}

var cnt int


func CommsLoop(conn net.Conn) error {
	var err error
	var cc byte
	cyc := 0
	cnt = 0

	for {

		// 0x88 = ping, 0x51 = start of data message.
		cc,err = Expect(conn, []byte{0x51, 0x88})
		if err != nil {
			return errors.Wrap(err, "initial expect")			// this includes the universal 2 minute watchdog/tarpit timeout.
		}

		if cc == 0x88 {
			fmt.Println("P")
			// fmt.Println("\ngot ping, continuing")
			continue
		}
		// fmt.Println("\ngot start of message.")

		// 0x69 = int-key = BMQ-LL
		cc,err = Expect(conn, []byte{0x69})
		if err != nil {
			return errors.Wrap(err, "not BMQ-LL message")
		}

		// fmt.Println("Woo, BMQ-LL message!")
		// Data len uvarint is next

		var dataLen int											// 0 by default
		dataLen, err = ReadUvarint(conn)						// we know hasData is on, so
		if err != nil {
			return errors.Wrap(err, "datalen ReadUvarint")
		}
		fmt.Println("Datalen: ",dataLen)
		// todo: cap dataLen

		// make a buffer that big, read that many bytes into it, then do stuff (b3 decode?) with the buffer. SS
		buf := make([]byte, dataLen)							// read reads up to the LEN of the slice not the CAP
		// buf len and dataLen are the same.
		// Loop-read until its full.
		for nread := 0; nread < dataLen; {
			// re-up the timeout  (lol go deadlines lol)
			if err = conn.SetReadDeadline(time.Now().Add(TIMEOUT)); err != nil {
				return errors.Wrap(err, "set timout")
			}
			// fmt.Println("\ncalling read, nread ",nread,"  buf len ",len(buf))
			n,nerr := conn.Read(buf[nread:])
			// fmt.Println("n    = ",n)
			// fmt.Println("nerr = ",nerr)
			if nerr != nil {
				return errors.Wrap(nerr, "buffer read")
			}
			if n == 0 {
				fmt.Println("read 0")
				break
			}
			//fmt.Print(Hexdump(buf, dataLen), "\n")
			nread += n
			// fmt.Println("nread now ",nread)
		}
		fmt.Println("Rx ",len(buf)," bytes successfully")
		// fmt.Printf("%d",cyc)
		cyc++
		if cyc > 9 {
			cyc = 0
		}
		cnt++

		// Currently we only process received BMQ-LL frames.
		frame := BMQLLFrame{}

		err = FillStructFromB3Buffer(buf, dataLen, &frame)
		if err != nil {
			return errors.Wrap(err, "filling bmqll struct")
		}

		// Pass it to app. (maybe send it out a channel)
		err = FrameReceived(frame)
		if err != nil {
			return errors.Wrap(err, "processing frame")
		}

	}
}

type BMQLLFrame struct {
	Cmd string `b3.type:"UTF8" b3.tag:"1"`
	Dat []byte `b3.type:"BYTES" b3.tag:"2"`
}

func FillStructFromB3Buffer(buf []byte, dataLen int, destStructPtr interface{}) error {
	//var bytesUsed int


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

	index := 0
	for index < len(buf) {
		hdr, bytesUsed, err := DecodeHeader(buf[index:])
		if err != nil {
			return errors.Wrap(err, "fillstruct decode header fail")
		}
		index += bytesUsed



	}



	// ==============================================================================================
	// We're passing slices. decode_header is special, it gets the [x:] rest-of-buf,
	// everything else gets [x:y] because sizes are KNOWN for everything else.
	// ==============================================================================================

	// Q: if we decode a header and there is an error, we cannot proceed yes?
	// A: correct because we can't be certain the dataLen in the header is legit because the header failed to decode.



    // for each b3 item in buf -
	// decode item header, get b3 data type, b3 key/tag, and datalen

	// (use reflect to) LOOK UP key/tag in struct with a loop. locate tag in dest struct, (ignore if not found)

	// ensure buf b3 data type we have, matches struct tag b3 data type (error if not)

	// do b3 codec lookup on data type, use b3 codec to create go concrete variable from data buf slice
	// (b3 codec universal function returns us an interface{} so that we can get our reflect.Value for Set() easily)
	// use .Set to insert go concrete variable into struct.


	// ----------- notes ----------

	// .Field panics if not struct, or if blah is out of range.
	// Reflect sure panics a lot.

	// we can call Set if we have a relect.Value to set into the struct,
	// which we will because b3 codec decode universal function is gonna return us an interface{}
	// Because if we're gonna be generic, we have to be generic on BOTH sides.

	// can only call Elem() on pointers (to get the writeable 'struct')
	// if you call Elem on a non-pointer it panics.
	// Actually "It panics if the type's Kind is not Array, Chan, Map, Ptr, or Slice."

	// we have to do a Kind check first.

	// lets make sure we can extract those struct tags too (see playground.go)

	return nil
}



// in go, strings are already utf8 []bytes really.
// go only utf8-decodes in 2 places 1) for i,r := range s (yielding runes), 2) casting []rune(s).
// In those instances invalide utf8 is replaces with U+FFFD (65533 utf8.RuneError) and the ops *do not crash*.
// https://stackoverflow.com/questions/34861479/how-to-detect-when-bytes-cant-be-converted-to-string-in-go
func TmpB3DecodeUTF8(buf []byte) (interface{}, error) {
	return string(buf),nil
}
func TmpB3DecodeBytes(buf []byte) (interface{}, error) {
	return buf,nil											// a no-op but interface{} is returned.
}


func FrameReceived(frame BMQLLFrame) error {
	fmt.Println("Bmq LL frame received! ",frame)
	return nil
}


func must(err error) {
	if err != nil {
		log.Fatalln("must >>>",err)
	}
}

func main() {
	//defer profile.Start().Stop()
	fmt.Println("Golang side")
	if bits.UintSize != 64 {
		panic("            **** Not in a 64bit mode! ( set GOARCH=amd64 ) ***")
	}

	go func() {
		for {
			fmt.Println("cnt ",cnt)
			cnt = 0
			time.Sleep(time.Second)
		}
	}()

	ConnectLoop()
	fmt.Println("Done.")
}


/*
	// "Note that the type assertion err.(net.Error) will correctly handle the nil case and return false for the ok value
	//       if nil is returned as the error, short-circuiting the Timeout check."
	if e, ok := err.(net.Error); ok && e.Timeout() {
		fmt.Println("Read cbyte timed out")
		continue
	} else if err != nil {
		fmt.Println("Read cbyte nontimeout error")
		return errors.Wrap(err, "read cbyte")
	}
	fmt.Println("Read cbyte success")
*/

/*
	for {
		fmt.Println("awaiting read")
		n, err := conn.Read(buf)
		fmt.Println("back from read")
		must(err)

		fmt.Println("Read ", n, " bytes")
		if n > 0 {
			fmt.Print(Hexdump(buf, n), "\n")

		}

	}
*/

/*
func CommsLoopWillBecomeItemHeaderDecode(conn net.Conn) error {
	// Returning nil shuts down, returning with an error causes a reconnect.
	var err error

	for {
		// Await/receive control byte
		if err = conn.SetReadDeadline(time.Now().Add(1*time.Second)); err != nil {
			return err
		}
		cbuf := make([]byte, 1)						// single byte buffer
		_,err = conn.Read(cbuf)					// read 1 byte

		if e, ok := err.(net.Error); ok && e.Timeout() {   // "Note that the type assertion err.(net.Error) will correctly handle the nil case and return false for the ok value if nil is returned as the error, short-circuiting the Timeout check."
			fmt.Println("Read cbyte timed out")
			continue
		} else if err != nil {
			fmt.Println("Read cbyte nontimeout error")
			return errors.Wrap(err, "read cbyte")
		}
		fmt.Println("Read cbyte success")

		cbyte := cbuf[0]

		// --- Validate cbyte ---
		// Policy: BMQ framing: frame must have-data, and not be null.
		isNull  := cbyte & 0x80 == 0x80
		if isNull {
			return errors.New("invalid frame (is-null set)")
		}
		hasData := cbyte & 0x40 == 0x40
		if !hasData {
			return errors.New("invalid frame (has-data not set)")
		}

		// --- Data type ---
		// Policy: BMQ framing: only DICT data type allowed.  (??)
		dataType := int(cbyte & 0x0f)					// base data type
		if dataType != 1 {								// must be B3_COMPOSITE_DICT
			return errors.New("invalid frame (data type must be dict)")
		}

		// --- Key ---
		// Policy: BMQ framing: only int (uvarint) keys. key is mandatory & how we signal inner proto. BMQ-LL is 0x69.
		keyTypeBits := cbyte & 0x30
		if keyTypeBits != 0x10 {						// UVARINT
			return errors.New("not an int key")
		}
		var key int
		key, err = ReadUvarint(conn)
		if err != nil {
			return errors.Wrap(err, "read int key")
		}
		if key != 0x69 {									// Policy: only support BMQLL at this time.
			return errors.New("Not a BMQ-LL frame (key 0x69 expected)")
		}

		// --- Null & data len ---
		var dataLen int										// 0 by default
		dataLen, err = ReadUvarint(conn)					// we know hasData is on, so
		if err != nil {
			return errors.Wrap(err, "read data len")
		}

		// === Now we have the data length, read the rest of the message ===
		// --- Bail if it takes too long (ie we are being tarpitted)     ---


	}
	return nil
}
 */
