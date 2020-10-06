package b3

import (
	"encoding/hex"
	"log"
	"strings"
)

// Takes strings like "11 22 aa bb", returns byte buffer.
// For making the tests more readable, basically.

func SBytes(bytesStr string) []byte {
	bytesStr = strings.ReplaceAll(bytesStr, " ", "") // chop spaces
	buf, err := hex.DecodeString(bytesStr)
	if err != nil {
		log.Fatal(err)
	}
	return buf
}

