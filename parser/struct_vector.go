package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Core/Public/Math/Vector.h#L29
type FVector struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

func (parser *PakParser) ReadFVector() *FVector {
	return &FVector{
		X: parser.ReadFloat32(),
		Y: parser.ReadFloat32(),
		Z: parser.ReadFloat32(),
	}
}
