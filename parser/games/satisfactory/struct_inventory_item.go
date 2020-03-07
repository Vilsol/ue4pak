package satisfactory

import "github.com/Vilsol/ue4pak/parser"

type FInventoryItem struct {
	ItemClass int32 `json:"is_valid"`
	ItemState int32 `json:"is_valid"`
}

func ReadFInventoryItem(parser *parser.PakParser) *FInventoryItem {
	return &FInventoryItem{
		ItemClass: parser.ReadInt32(),
		ItemState: parser.ReadInt32(),
	}
}
