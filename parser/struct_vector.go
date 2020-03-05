package parser

type FVector struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

func (parser *PakParser) ReadFVector() *FVector {
	return &FVector{
		X: parser.ReadFloat32(),
		Y: parser.ReadFloat32(),
		Z: parser.ReadFloat32(),
	}
}
