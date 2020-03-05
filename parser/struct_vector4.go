package parser

type FVector4 struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
	W float32 `json:"w"`
}

func (parser *PakParser) ReadFVector4() *FVector4 {
	return &FVector4{
		X: parser.ReadFloat32(),
		Y: parser.ReadFloat32(),
		Z: parser.ReadFloat32(),
		W: parser.ReadFloat32(),
	}
}
