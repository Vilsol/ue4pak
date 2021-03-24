package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fatih/color"
	"math"
	"strings"
	"unicode/utf16"
)

func Int16(b []byte) int16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return int16(b[0]) | int16(b[1])<<8
}

func Int32(b []byte) int32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
}

func Int64(b []byte) int64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return int64(b[0]) | int64(b[1])<<8 | int64(b[2])<<16 | int64(b[3])<<24 |
		int64(b[4])<<32 | int64(b[5])<<40 | int64(b[6])<<48 | int64(b[7])<<56
}

func Float32(b []byte) float32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return math.Float32frombits(binary.LittleEndian.Uint32(b[:4]))
}

func safeChar(char byte) string {
	if char <= 0x1F {
		return "."
	}

	return string(char)
}

func HexDump(data []byte) string {
	return HexDumpWidth(data, 32)
}

func HexDumpWidth(data []byte, perRow int) string {
	result := ""

	rows := int(math.Ceil(float64(len(data)) / float64(perRow)))

	lengthChars := 0
	stringChars := uint32(0)

	for i := 0; i < rows; i++ {
		hexSide := ""
		charSide := ""

		for k := 0; k < perRow && k < len(data[i*perRow:]); k++ {
			offset := i*perRow + k

			if lengthChars == 0 && stringChars == 0 && IsLengthPrefixedNullTerminatedString(data[offset:]) {
				lengthChars = 4
				stringChars = binary.LittleEndian.Uint32(data[offset:])
				// fmt.Println(string(data[offset+4:4+uint32(offset)+binary.LittleEndian.Uint32(data[offset:])]))
			}

			hexTemp := fmt.Sprintf("%#-4x", data[offset]) + " "
			charTemp := safeChar(data[offset])

			switch {
			case lengthChars > 0:
				hexSide += color.BlueString(hexTemp)
				charSide += color.BlueString(charTemp)
				lengthChars--
			case stringChars > 0:
				hexSide += color.GreenString(hexTemp)
				charSide += color.GreenString(charTemp)
				stringChars--
			default:
				hexSide += hexTemp
				charSide += charTemp
			}
		}

		result += fmt.Sprintf("%-#6x: %s", i*perRow, hexSide)
		if len(data[i*perRow:]) < perRow && len(data) > perRow {
			result += strings.Repeat(" ", (perRow-len(data[i*perRow:]))*5)
		}
		result += fmt.Sprintf("%s\n", charSide)
	}

	return result
}

func IsLengthPrefixedNullTerminatedString(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	strLen := binary.LittleEndian.Uint32(data)

	if strLen <= 1 {
		return false
	}

	if strLen > uint32(len(data)-4) {
		return false
	}

	if data[4+strLen-1] != 0 {
		return false
	}

	for i := 0; i < int(strLen-1); i++ {
		if data[4+i] <= 0x1F || data[4+i] >= 0x7F {
			return false
		}
	}

	return true
}

func DecodeUtf16(b []byte) string {
	ints := make([]uint16, len(b)/2)
	binary.Read(bytes.NewReader(b), binary.LittleEndian, &ints)
	return string(utf16.Decode(ints))
}
