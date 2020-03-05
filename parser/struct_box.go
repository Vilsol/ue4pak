package parser

type FBox struct {
	IsValid uint8    `json:"is_valid"`
	Min     *FVector `json:"min"`
	Max     *FVector `json:"max"`
}

func (parser *PakParser) ReadFBox() *FBox {
	return &FBox{
		IsValid: parser.Read(1)[0],
		Min:     parser.ReadFVector(),
		Max:     parser.ReadFVector(),
	}
}
