package parser

import (
	"encoding/binary"
	"github.com/Vilsol/ue4pak/utils"
	"math"
)

func (parser *PakParser) ReadString() string {
	stringLength := parser.ReadInt32()

	if stringLength == 0 {
		return ""
	}

	if stringLength < 0 {
		stringLength = (stringLength * -1) * 2
		return utils.DecodeUtf16(parser.Read(stringLength))
	}

	return string(parser.Read(stringLength))
}

func (parser *PakParser) ReadStringNull() string {
	result := make([]byte, 0)

	for {
		b := parser.Read(1)[0]
		if b == 0x00 {
			break
		} else {
			result = append(result, b)
		}
	}

	return string(result)
}

func (parser *PakParser) ReadFloat32() float32 {
	value := math.Float32frombits(parser.ReadUint32())
	assertFloat32IsFinite(value)
	return value
}

func (parser *PakParser) ReadInt32() int32 {
	return utils.Int32(parser.Read(4))
}

func (parser *PakParser) ReadInt64() int64 {
	return utils.Int64(parser.Read(8))
}

func (parser *PakParser) ReadUint16() uint16 {
	return binary.LittleEndian.Uint16(parser.Read(2))
}

func (parser *PakParser) ReadUint32() uint32 {
	return binary.LittleEndian.Uint32(parser.Read(4))
}

func (parser *PakParser) ReadUint64() uint64 {
	return binary.LittleEndian.Uint64(parser.Read(8))
}
