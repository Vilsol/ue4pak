package parser

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
