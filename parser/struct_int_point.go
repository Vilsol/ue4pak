package parser

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
