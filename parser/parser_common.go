package parser

import (
	"math"
	"strings"
)

func (parser *PakParser) ReadFGenerationInfo() *FGenerationInfo {
	return &FGenerationInfo{
		ExportCount: parser.ReadInt32(),
		NameCount:   parser.ReadInt32(),
	}
}

func (parser *PakParser) ReadFEngineVersion() *FEngineVersion {
	return &FEngineVersion{
		Major:      parser.ReadUint16(),
		Minor:      parser.ReadUint16(),
		Patch:      parser.ReadUint16(),
		ChangeList: parser.ReadUint32(),
		Branch:     parser.ReadString(),
	}
}

func (parser *PakParser) ReadFCompressedChunk() *FCompressedChunk {
	return &FCompressedChunk{
		UncompressedOffset: parser.ReadInt32(),
		UncompressedSize:   parser.ReadInt32(),
		CompressedOffset:   parser.ReadInt32(),
		CompressedSize:     parser.ReadInt32(),
	}
}

func (parser *PakParser) ReadFName(names []*FNameEntrySerialized) string {
	index := parser.ReadUint32()
	// Instance ID
	parser.Read(4)
	return names[index].Name
}

func (parser *PakParser) ReadFPackageIndex(imports []*FObjectImport, exports []*FObjectExport) *FPackageIndex {
	return parser.ReadFPackageIndexInt(parser.ReadInt32(), imports, exports)
}

func (parser *PakParser) ReadFPackageIndexInt(index int32, imports []*FObjectImport, exports []*FObjectExport) *FPackageIndex {
	if index == 0 {
		// TODO Values of 0 indicate that this resource represents a top-level UPackage object (the linker's LinkerRoot). Serialized
		return &FPackageIndex{
			Index:     index,
			Reference: nil,
		}
	}

	if index < 0 {
		correctedIndex := index*-1 - 1
		if correctedIndex >= 0 && correctedIndex < int32(len(imports)) {
			return &FPackageIndex{
				Index:     index,
				Reference: imports[index*-1-1],
			}
		}

		return &FPackageIndex{
			Index:     index,
			Reference: nil,
		}
	}

	if index-1 < int32(len(exports)) {
		return &FPackageIndex{
			Index:     index - 1,
			Reference: exports[index-1],
		}
	}

	return nil
}

func (parser *PakParser) ReadFText() *FText {
	flags := parser.ReadUint32()
	historyType := int8(parser.Read(1)[0])

	text := FText{
		Flags:       flags,
		HistoryType: historyType,
	}

	if historyType != 0 {
		return &text
	}

	text.Namespace = parser.ReadString()
	text.Key = parser.ReadString()
	text.SourceString = parser.ReadString()

	return &text
}

func (parser *PakParser) ReadFPropertyTagLoop(uAsset *FPackageFileSummary) []*FPropertyTag {
	properties := make([]*FPropertyTag, 0)

	for {
		property := parser.ReadFPropertyTag(uAsset, true, 0)

		if property == nil {
			break
		}

		properties = append(properties, property)
	}

	return properties
}

func d(n int) string {
	return strings.Repeat("  ", n)
}

func assertFloat32IsFinite(n float32) {
	value := float64(n)
	if math.IsNaN(value) {
		panic("Expected a float32, but received NaN")
	}
	if math.IsInf(value, 0) {
		panic("Expected a float32, but received inf")
	}
}
