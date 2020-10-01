package main

import (
"fmt"
"os"
"strings"

"golang.org/x/sys/windows"
)

// --- Globals ---
var gColorOK bool = false

func TryEnableWin10TerminalAnsi() { // is best-effort, no error return
	// Win10 has support for ANSI sequences in its cmd terminal windows finally. This works in cmd and clink.
	// On non-win10 OSes, the SetConsoleMode call returns failure (0), on win10 it returns sucess (1)
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32
	_ = windows.GetConsoleMode(stdout, &originalMode)
	if err := windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err == nil {
		// fmt.Println("Set Mode returned ok")
		gColorOK = true
	}
}

// python chops its way through its byte buffer
// we want to slice our way through the source slice ??
// using for-range on strings gives us runes, but we"re using []byte here so should get bytes
// strings.Builder is like a bytes.Buffer, but for strings i guess.

const LENGTH = 16

func DotStrFromBytes(src []byte) string {
	var out strings.Builder
	for _, c := range (src) {
		if c > 31 && c < 127 {
			out.WriteByte(c)
		} else {
			out.WriteString(".")
		}
	}
	return out.String()
}

var cols = map[string]string{
	"rst": "\x1b[0m", "gry": "\x1b[2;37m", "red": "\x1b[31m", "yel": "\x1b[0;33m", "byel": "\x1b[1;33m",
	"grn": "\x1b[0;32m", "bgrn": "\x1b[1;32m", "blu": "\x1b[0;34m", "bblu": "\x1b[34;1m",
}

func ByteToColorCode(x uint8) (col string) {
	switch x {
	case 13, 10: // CR LF
		col = "bgrn"
	case 32: // Space
		col = "bblu"
	default:
		if x > 32 && x < 127 { // printable
			col = "grn"
		} else {
			col = "rst" // unprintable
		}
	}
	return cols[col]
}

// maps return the 'zero type' if key not found.
// Can also use  ret,ok := cols[col]   - ret will still be the 'zero type' if ok is false.
// fmt.Println("ok ",ok,"  len ret ",len(ret))
// return ret

// The answer is simple: It is common to *compute* an index and such computations tend to underflow much too easy if done in unsigned integers.
// https://stackoverflow.com/questions/39088945/why-does-len-returned-a-signed-value

// really want srclen to be an int, even though that makes no sense, because len() returns an int.

// GO doesn't have optional function arguments. See evernote for workarounds.
// Simplest workaround is to have Hexdump() calls HexdumpWithPrefix()

func Hexdump(src []byte, srclen int) string { // best-effort, no error return currently
	if srclen < 1 {
		return " *** HEXDUMP ERROR len < 1 *** "		// can happen because of go's zero-value ethos
	}

	var lines []string

	for offset := 0; offset < srclen; offset += LENGTH {
		var line strings.Builder

		offset_end := offset + LENGTH
		if offset_end >= srclen {
			offset_end = srclen
		}

		s := src[offset:offset_end]

		nccl := 0
		oldCol := ""
		colCode := ""
		line.WriteString(fmt.Sprintf("%04X ", offset)) // address

		for _, x := range s {
			if gColorOK == true { // Do color coding for hex bytes
				colCode = ByteToColorCode(x)
				if colCode != oldCol {
					line.WriteString(colCode)
					oldCol = colCode
				}
			}
			line.WriteString(fmt.Sprintf(" %02x", x)) // hex bytes
			nccl += 3
		}
		if gColorOK == true {
			line.WriteString(cols["rst"])
		}

		line.WriteString(strings.Repeat(" ", (LENGTH*3)-nccl)) // pad
		line.WriteString("  ")
		line.WriteString(DotStrFromBytes(s)) // dot-str bytes
		lines = append(lines, line.String())
	}
	return strings.Join(lines, "\n")
}

func init() {
	TryEnableWin10TerminalAnsi()
}

// func main() {
// 	fmt.Println("i am hexdumps main")
// 	const FOO = "hello \x01\x02\x07\x0f\x12world\nThis is a drill\034\x67\x21\x08\x09\x10\x11\x12\x13\x14 this is a drill, good morning vietnam!"
// 	fmt.Print(HexdumpWithPrefix([]byte(FOO), len(FOO), "hello"))
// }
