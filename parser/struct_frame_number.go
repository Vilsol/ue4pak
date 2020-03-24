package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Core/Public/Misc/FrameNumber.h#L16
type FFrameNumber struct {
	Value int32 `json:"value"`
}

func (parser *PakParser) ReadFFrameNumber() *FFrameNumber {
	return &FFrameNumber{
		Value: parser.ReadInt32(),
	}
}
