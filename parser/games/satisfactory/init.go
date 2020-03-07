package satisfactory

import "github.com/Vilsol/ue4pak/parser"

func init() {
	parser.RegisterMapPropertyOverride("ChildrenAndRoads_34_758C9E0D4F09DAF4BBAD309358952A0A", &parser.MapProperty{
		KeyType:   "IntVector2D",
		ValueType: "MAMTree_RoadPoints",
	})

	parser.RegisterStructResolver("InventoryItem", func(parser *parser.PakParser, property *parser.StructProperty, size int32, uAsset *parser.FPackageFileSummary) interface{} {
		return ReadFInventoryItem(parser)
	})
}
