package parser

type FFrameNumber struct {
	Value int32 `json:"value"`
}

func (parser *PakParser) ReadFFrameNumber() *FFrameNumber {
	return &FFrameNumber{
		Value: parser.ReadInt32(),
	}
}
