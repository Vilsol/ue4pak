package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Core/Public/Math/IntVector.h#L14
type FIntVector struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
	Z int32 `json:"z"`
}

func (parser *PakParser) ReadFIntVector() *FIntVector {
	return &FIntVector{
		X: parser.ReadInt32(),
		Y: parser.ReadInt32(),
		Z: parser.ReadInt32(),
	}
}
