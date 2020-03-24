package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Engine/Classes/Curves/KeyHandle.h#L13
type FKeyHandle struct {
	Index int32 `json:"index"`
}

func (parser *PakParser) ReadFKeyHandle() *FKeyHandle {
	return &FKeyHandle{
		Index: parser.ReadInt32(),
	}
}
