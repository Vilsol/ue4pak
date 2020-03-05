package parser

type FColor struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}

func (parser *PakParser) ReadFColor() *FColor {
	return &FColor{
		R: parser.Read(1)[0],
		G: parser.Read(1)[0],
		B: parser.Read(1)[0],
		A: parser.Read(1)[0],
	}
}
