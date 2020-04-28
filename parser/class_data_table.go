package parser

// https://github.com/SatisfactoryModdingUE/UnrealEngine/blob/4.22-CSS/Engine/Source/Runtime/Engine/Classes/Engine/DataTable.h#L56
type UDataTable struct {
	Values map[string][]*FPropertyTag `json:"values"`
}

func (parser *PakParser) ReadUDataTable(uAsset *FPackageFileSummary) *UDataTable {
	// Unknown
	parser.Read(4)

	count := parser.ReadUint32()

	values := make(map[string][]*FPropertyTag)

	for i := uint32(0); i < count; i++ {
		name := parser.ReadFName(uAsset.Names)
		values[name] = parser.ReadFPropertyTagLoop(uAsset)
	}

	return &UDataTable{
		Values: values,
	}
}
