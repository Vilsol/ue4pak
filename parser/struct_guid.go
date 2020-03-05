package parser

type FGuid struct {
	A uint32 `json:"a"`
	B uint32 `json:"b"`
	C uint32 `json:"c"`
	D uint32 `json:"d"`
}

func (parser *PakParser) ReadFGuid() *FGuid {
	return &FGuid{
		A: parser.ReadUint32(),
		B: parser.ReadUint32(),
		C: parser.ReadUint32(),
		D: parser.ReadUint32(),
	}
}
