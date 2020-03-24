package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Core/Public/Math/Vector2D.h#L17
type FVector2D struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

func (parser *PakParser) ReadFVector2D() *FVector2D {
	return &FVector2D{
		X: parser.ReadFloat32(),
		Y: parser.ReadFloat32(),
	}
}
