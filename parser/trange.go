package parser

import (
	"fmt"
)

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Core/Public/Math/RangeBound.h#L34
type TRangeBound struct {
	BoundType uint8       `json:"bound_type"`
	Value     interface{} `json:"value"`
}

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Core/Public/Math/Range.h#L48
type TRange struct {
	LowerBound *TRangeBound `json:"lower_bound"`
	UpperBound *TRangeBound `json:"upper_bound"`
}

func (parser *PakParser) ReadTRange(t string) *TRange {
	return &TRange{
		LowerBound: parser.ReadTRangeBound(t),
		UpperBound: parser.ReadTRangeBound(t),
	}
}

func (parser *PakParser) ReadTRangeBound(t string) *TRangeBound {
	boundType := parser.Read(1)[0]
	var value interface{}

	switch t {
	case "int32":
		value = parser.ReadInt32()
		break
	default:
		panic(fmt.Sprintf("Unkown bound type: %s", t))
	}

	return &TRangeBound{
		BoundType: boundType,
		Value:     value,
	}
}
