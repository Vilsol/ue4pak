package parser

// https://github.com/EpicGames/UnrealEngine/blob/4.22/Engine/Source/Runtime/Core/Public/Math/Box.h#L17
type FBox struct {
	Min     *FVector `json:"min"`
	Max     *FVector `json:"max"`
	IsValid uint8    `json:"is_valid"`
}

func (parser *PakParser) ReadFBox() *FBox {
	return &FBox{
		Min:     parser.ReadFVector(),
		Max:     parser.ReadFVector(),
		IsValid: parser.Read(1)[0],
	}
}
