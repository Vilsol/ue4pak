package parser

import "log"

type TRangeBound struct {
	BoundType uint8       `json:"bound_type"`
	Value     interface{} `json:"value"`
}

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
		log.Fatalf("Unkown bound type: %s", t)
		break
	}

	return &TRangeBound{
		BoundType: boundType,
		Value:     value,
	}
}
