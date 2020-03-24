package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Core/Public/Math/IntPoint.h#L16
type FIntPoint struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

func (parser *PakParser) ReadFIntPoint() *FIntPoint {
	return &FIntPoint{
		X: parser.ReadInt32(),
		Y: parser.ReadInt32(),
	}
}
