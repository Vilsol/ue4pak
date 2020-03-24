package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Core/Public/Math/Color.h#L31
type FLinearColor struct {
	R float32 `json:"r"`
	G float32 `json:"g"`
	B float32 `json:"b"`
	A float32 `json:"a"`
}

func (parser *PakParser) ReadFLinearColor() *FLinearColor {
	return &FLinearColor{
		R: parser.ReadFloat32(),
		G: parser.ReadFloat32(),
		B: parser.ReadFloat32(),
		A: parser.ReadFloat32(),
	}
}
