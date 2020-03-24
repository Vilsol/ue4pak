package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Engine/Classes/Materials/MaterialExpression.h#L22
type FExpressionInput struct {
	Expression     uint32 `json:"expression"`
	ExpressionName string `json:"expression_name"`
	Mask           int32  `json:"mask"`
	MaskR          int32  `json:"mask_r"`
	MaskG          int32  `json:"mask_g"`
	MaskB          int32  `json:"mask_b"`
	MaskA          int32  `json:"mask_a"`
	OutputIndex    int32  `json:"output_index"`
}

func (parser *PakParser) ReadFExpressionInput(names []*FNameEntrySerialized) *FExpressionInput {
	return &FExpressionInput{
		Expression:     parser.ReadUint32(),
		ExpressionName: parser.ReadFName(names),
		Mask:           parser.ReadInt32(),
		MaskR:          parser.ReadInt32(),
		MaskG:          parser.ReadInt32(),
		MaskB:          parser.ReadInt32(),
		MaskA:          parser.ReadInt32(),
		OutputIndex:    parser.ReadInt32(),
	}
}
