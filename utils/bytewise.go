package utils

import (
	"encoding/binary"
	"fmt"
	"github.com/fatih/color"
	"math"
	"strconv"
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

func safeChar(char byte) string {
	if char <= 0x1F {
		return "."
	}

	return string(char)
}

func HexDump(data []byte) string {
	result := ""

	perRow := 32
	rows := int(math.Ceil(float64(len(data)) / float64(perRow)))

	rowWidth := perRow * 5
	if len(data) < perRow {
		rowWidth = len(data) * 5
	}

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
			charTemp := fmt.Sprintf("%s", safeChar(data[offset]))

			if lengthChars > 0 {
				hexSide += color.BlueString(hexTemp)
				charSide += color.BlueString(charTemp)
				lengthChars--
			} else if stringChars > 0 {
				hexSide += color.GreenString(hexTemp)
				charSide += color.GreenString(charTemp)
				stringChars--
			} else {
				hexSide += hexTemp
				charSide += charTemp
			}
		}

		result += fmt.Sprintf("%-#6x: %-"+strconv.Itoa(rowWidth)+"s%s\n", i*perRow, hexSide, charSide)
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
