package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Core/Public/Math/Box2D.h#L14
type FBox2D struct {
	IsValid uint8      `json:"is_valid"`
	Min     *FVector2D `json:"min"`
	Max     *FVector2D `json:"max"`
}

func (parser *PakParser) ReadFBox2D() *FBox2D {
	return &FBox2D{
		IsValid: parser.Read(1)[0],
		Min:     parser.ReadFVector2D(),
		Max:     parser.ReadFVector2D(),
	}
}
