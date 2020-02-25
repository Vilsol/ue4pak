package parser

import (
	"encoding/binary"
	"fmt"
	"github.com/Vilsol/ue4pak/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"math"
	"strings"
)

type PakParser struct {
	reader  PakReader
	tracker *readTracker
	preload []byte
}

type readTracker struct {
	child     *readTracker
	bytesRead int32
}

func (tracker *readTracker) Increment(n int32) {
	tracker.bytesRead += n

	if tracker.child != nil {
		tracker.child.Increment(n)
	}
}

func NewParser(reader PakReader) *PakParser {
	return &PakParser{
		reader: reader,
	}
}

func (parser *PakParser) ProcessPak(parseFile func(string) bool) ([]*PakEntrySet, *PakFile) {
	pak := parser.Parse()

	results := make([]*PakEntrySet, 0)

	summaries := make(map[string]*FPackageFileSummary, 0)

	// First pass, parse summaries
	for j, record := range pak.Index.Records {
		trimmed := strings.Trim(record.FileName, "\x00")

		if parseFile != nil {
			if !parseFile(trimmed) {
				continue
			}
		}

		if strings.HasSuffix(trimmed, "uasset") {
			offset := record.FileOffset + pak.Footer.HeaderSize()
			log.Infof("Reading Record: %d [%x-%x]: %s\n", j, offset, offset+record.FileSize, trimmed)
			summaries[trimmed[0:strings.Index(trimmed, ".uasset")]] = record.ReadUAsset(pak, parser)
			summaries[trimmed[0:strings.Index(trimmed, ".uasset")]].Record = record
		}
	}

	// Second pass, parse exports
	for j, record := range pak.Index.Records {
		trimmed := strings.Trim(record.FileName, "\x00")

		if parseFile != nil {
			if !parseFile(trimmed) {
				continue
			}
		}

		if strings.HasSuffix(trimmed, "uexp") {
			summary, ok := summaries[trimmed[0:strings.Index(trimmed, ".uexp")]]

			offset := record.FileOffset + pak.Footer.HeaderSize()

			if !ok {
				log.Errorf("Unable to read record. Missing uasset: %d [%x-%x]: %s\n", j, offset, offset+record.FileSize, trimmed)
				continue
			}

			log.Infof("Reading Record: %d [%x-%x]: %s\n", j, offset, offset+record.FileSize, trimmed)

			exports := record.ReadUExp(pak, parser, summary)

			exportSet := make([]PakExportSet, len(exports))

			i := 0
			for export, properties := range exports {
				exportSet[i] = PakExportSet{
					Export:     export,
					Properties: properties,
				}
				i++
			}

			results = append(results, &PakEntrySet{
				ExportRecord: record,
				Summary:      summary,
				Exports:      exportSet,
			})
		}
	}

	return results, pak
}

func (parser *PakParser) Parse() *PakFile {
	// Find magic number
	magicOffset := int64(-44)

	for {
		parser.Seek(magicOffset, 2)

		magicArray := parser.Read(4)

		if magicArray[0] == 0xE1 && magicArray[1] == 0x12 && magicArray[2] == 0x6F && magicArray[3] == 0x5A {
			break
		}

		magicOffset -= 1

		if magicOffset < -1024 {
			log.Fatal("Could not find magic bytes in pak!")
		}
	}

	// Seek and read the footer of the file
	parser.Seek(magicOffset, 2)
	footer := parser.Read(int32(magicOffset * -1))

	pakFooter := &FPakInfo{
		Magic:         binary.LittleEndian.Uint32(footer[0:4]),
		Version:       binary.LittleEndian.Uint32(footer[4:8]),
		IndexOffset:   binary.LittleEndian.Uint64(footer[8:16]),
		IndexSize:     binary.LittleEndian.Uint64(footer[16:24]),
		IndexSHA1Hash: footer[24:44],
	}

	// Seek and read the index of the file
	parser.Seek(int64(pakFooter.IndexOffset), 0)

	mountPoint := parser.ReadString()
	recordCount := parser.ReadUint32()

	pakIndex := &FPakIndex{
		MountPoint: mountPoint,
		Records:    make([]*FPakEntry, recordCount),
	}

	for i := 0; i < len(pakIndex.Records); i++ {
		pakIndex.Records[i] = &FPakEntry{}

		pakIndex.Records[i].FileName = parser.ReadString()
		pakIndex.Records[i].FileOffset = parser.ReadUint64()
		pakIndex.Records[i].FileSize = parser.ReadUint64()
		pakIndex.Records[i].UncompressedSize = parser.ReadUint64()

		if pakFooter.Version >= 8 {
			pakIndex.Records[i].CompressionMethod = uint32(parser.Read(1)[0])
		} else {
			pakIndex.Records[i].CompressionMethod = parser.ReadUint32()
		}

		if pakFooter.Version <= 1 {
			pakIndex.Records[i].Timestamp = parser.ReadUint64()
		}

		pakIndex.Records[i].DataSHA1Hash = parser.Read(20)

		if pakFooter.Version >= 3 {
			if pakIndex.Records[i].CompressionMethod != 0 {
				blockCount := parser.ReadUint32()

				pakIndex.Records[i].CompressionBlocks = make([]*FPakCompressedBlock, blockCount)

				for j := 0; j < len(pakIndex.Records[i].CompressionBlocks); j++ {
					pakIndex.Records[i].CompressionBlocks[j] = &FPakCompressedBlock{
						StartOffset: parser.ReadUint64(),
						EndOffset:   parser.ReadUint64(),
					}
				}
			}

			pakIndex.Records[i].IsEncrypted = parser.Read(1)[0] > 0
			pakIndex.Records[i].CompressionBlockSize = parser.ReadUint32()
		}

		if pakFooter.Version == 4 {
			// TODO ???
		}
	}

	return &PakFile{
		Footer: pakFooter,
		Index:  pakIndex,
	}
}

func (parser *PakParser) TrackRead() {
	parser.tracker = &readTracker{
		child: parser.tracker,
	}
}

func (parser *PakParser) UnTrackRead() {
	if parser.tracker != nil {
		parser.tracker = parser.tracker.child
	}
}

func (parser *PakParser) Seek(offset int64, whence int) (ret int64, err error) {
	parser.preload = nil
	return parser.reader.Seek(offset, whence)
}

func (parser *PakParser) Preload(n int32) {
	if viper.GetBool("NoPreload") {
		return
	}

	buffer := make([]byte, n)
	read, err := parser.reader.Read(buffer)

	if err != nil {
		panic(err)
	}

	if int32(read) < n {
		panic(fmt.Sprintf("End of stream: %d < %d", read, n))
	}

	if parser.preload != nil && len(parser.preload) > 0 {
		parser.preload = append(parser.preload, buffer...)
	} else {
		parser.preload = buffer
	}
}

func (parser *PakParser) Read(n int32) []byte {
	toRead := n
	buffer := make([]byte, toRead)

	if parser.preload != nil && len(parser.preload) > 0 {
		copied := copy(buffer, parser.preload)
		parser.preload = parser.preload[copied:]
		toRead = toRead - int32(copied)
	}

	if toRead > 0 {
		read, err := parser.reader.Read(buffer[n-toRead:])

		if err != nil {
			panic(err)
		}

		if int32(read) < toRead {
			panic(fmt.Sprintf("End of stream: %d < %d", read, toRead))
		}
	}

	if parser.tracker != nil {
		parser.tracker.Increment(n)
	}

	return buffer
}

func (parser *PakParser) ReadString() string {
	stringLength := parser.ReadInt32()

	if stringLength == 0 {
		return ""
	}

	if stringLength < 0 {
		stringLength = (stringLength * -1) * 2
		return utils.DecodeUtf16(parser.Read(stringLength))
	}

	return string(parser.Read(stringLength))
}

func (parser *PakParser) ReadFloat32() float32 {
	return math.Float32frombits(parser.ReadUint32())
}

func (parser *PakParser) ReadInt32() int32 {
	return utils.Int32(parser.Read(4))
}

func (parser *PakParser) ReadInt64() int64 {
	return utils.Int64(parser.Read(8))
}

func (parser *PakParser) ReadUint16() uint16 {
	return binary.LittleEndian.Uint16(parser.Read(2))
}

func (parser *PakParser) ReadUint32() uint32 {
	return binary.LittleEndian.Uint32(parser.Read(4))
}

func (parser *PakParser) ReadUint64() uint64 {
	return binary.LittleEndian.Uint64(parser.Read(8))
}

func (parser *PakParser) ReadFGuid() *FGuid {
	data := parser.Read(16)
	return &FGuid{
		A: binary.LittleEndian.Uint32(data),
		B: binary.LittleEndian.Uint32(data[4:]),
		C: binary.LittleEndian.Uint32(data[8:]),
		D: binary.LittleEndian.Uint32(data[12:]),
	}
}

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

	if historyType != 0 {
		return &FText{
			Flags:       flags,
			HistoryType: historyType,
		}
	}

	return &FText{
		Flags:        flags,
		HistoryType:  historyType,
		Namespace:    parser.ReadString(),
		Key:          parser.ReadString(),
		SourceString: parser.ReadString(),
	}
}

func (record *FPakEntry) ReadUAsset(pak *PakFile, parser *PakParser) *FPackageFileSummary {
	// Skip UE4 pak header
	// TODO Find out what's in the pak header
	headerSize := int64(pak.Footer.HeaderSize())

	parser.Seek(headerSize+int64(record.FileOffset), 0)
	parser.Preload(int32(record.FileSize))

	tag := parser.ReadInt32()
	legacyFileVersion := parser.ReadInt32()
	legacyUE3Version := parser.ReadInt32()
	fileVersionUE4 := parser.ReadInt32()
	fileVersionLicenseeUE4 := parser.ReadInt32()

	// TODO custom_version_container: Vec<FCustomVersion>
	parser.Read(4)

	totalHeaderSize := parser.ReadInt32()
	folderName := parser.ReadString()
	packageFlags := parser.ReadUint32()
	nameCount := parser.ReadUint32()
	nameOffset := parser.ReadInt32()
	gatherableTextDataCount := parser.ReadInt32()
	gatherableTextDataOffset := parser.ReadInt32()
	exportCount := parser.ReadUint32()
	exportOffset := parser.ReadInt32()
	importCount := parser.ReadUint32()
	importOffset := parser.ReadInt32()
	dependsOffset := parser.ReadInt32()
	stringAssetReferencesCount := parser.ReadInt32()
	stringAssetReferencesOffset := parser.ReadInt32()
	searchableNamesOffset := parser.ReadInt32()
	thumbnailTableOffset := parser.ReadInt32()
	guid := parser.ReadFGuid()
	generationCount := parser.ReadUint32()

	generations := make([]*FGenerationInfo, generationCount)
	for i := uint32(0); i < generationCount; i++ {
		generations[i] = parser.ReadFGenerationInfo()
	}

	savedByEngineVersion := parser.ReadFEngineVersion()
	compatibleWithEngineVersion := parser.ReadFEngineVersion()
	compressionFlags := parser.ReadUint32()
	compressedChunkCount := parser.ReadUint32()

	compressedChunks := make([]*FCompressedChunk, compressedChunkCount)
	for i := uint32(0); i < compressedChunkCount; i++ {
		compressedChunks[i] = parser.ReadFCompressedChunk()
	}

	packageSource := parser.ReadUint32()
	additionalPackageCount := parser.ReadUint32()

	additionalPackagesToCook := make([]string, additionalPackageCount)
	for i := uint32(0); i < additionalPackageCount; i++ {
		additionalPackagesToCook[i] = parser.ReadString()
	}

	assetRegistryDataOffset := parser.ReadInt32()
	bulkDataStartOffset := parser.ReadInt32()
	worldTileInfoDataOffset := parser.ReadInt32()
	chunkCount := parser.ReadUint32()

	chunkIds := make([]int32, chunkCount)
	for i := uint32(0); i < chunkCount; i++ {
		chunkIds[i] = parser.ReadInt32()
	}

	// TODO unknown bytes
	parser.Read(4)

	preloadDependencyCount := parser.ReadInt32()
	preloadDependencyOffset := parser.ReadInt32()

	names := make([]*FNameEntrySerialized, nameCount)
	for i := uint32(0); i < nameCount; i++ {
		names[i] = &FNameEntrySerialized{
			Name:                  parser.ReadString(),
			NonCasePreservingHash: parser.ReadUint16(),
			CasePreservingHash:    parser.ReadUint16(),
		}
	}

	imports := make([]*FObjectImport, importCount)
	for i := uint32(0); i < importCount; i++ {
		imports[i] = &FObjectImport{
			ClassPackage: parser.ReadFName(names),
			ClassName:    parser.ReadFName(names),
			OuterIndex:   parser.ReadInt32(),
			ObjectName:   parser.ReadFName(names),
		}
	}

	exports := make([]*FObjectExport, exportCount)
	for i := uint32(0); i < exportCount; i++ {
		exports[i] = &FObjectExport{
			ClassIndex:                   parser.ReadFPackageIndex(imports, exports),
			SuperIndex:                   parser.ReadFPackageIndex(imports, exports),
			TemplateIndex:                parser.ReadFPackageIndex(imports, exports),
			OuterIndex:                   parser.ReadFPackageIndex(imports, exports),
			ObjectName:                   parser.ReadFName(names),
			Save:                         parser.ReadUint32(),
			SerialSize:                   parser.ReadInt64(),
			SerialOffset:                 parser.ReadInt64(),
			ForcedExport:                 parser.ReadInt32() != 0,
			NotForClient:                 parser.ReadInt32() != 0,
			NotForServer:                 parser.ReadInt32() != 0,
			PackageGuid:                  parser.ReadFGuid(),
			PackageFlags:                 parser.ReadUint32(),
			NotAlwaysLoadedForEditorGame: parser.ReadInt32() != 0,
			IsAsset:                      parser.ReadInt32() != 0,
			FirstExportDependency:        parser.ReadInt32(),
			SerializationBeforeSerializationDependencies: parser.ReadInt32() != 0,
			CreateBeforeSerializationDependencies:        parser.ReadInt32() != 0,
			SerializationBeforeCreateDependencies:        parser.ReadInt32() != 0,
			CreateBeforeCreateDependencies:               parser.ReadInt32() != 0,
		}
	}

	for _, objectImport := range imports {
		objectImport.OuterPackage = parser.ReadFPackageIndexInt(objectImport.OuterIndex, imports, exports)
	}

	// fmt.Println("UASSET LEFTOVERS:", len(fileData[offset:]))
	// fmt.Println(utils.HexDump(fileData[offset:]))

	// TODO Bunch of unknown bytes at the end

	return &FPackageFileSummary{
		Tag:                         tag,
		LegacyFileVersion:           legacyFileVersion,
		LegacyUE3Version:            legacyUE3Version,
		FileVersionUE4:              fileVersionUE4,
		FileVersionLicenseeUE4:      fileVersionLicenseeUE4,
		TotalHeaderSize:             totalHeaderSize,
		FolderName:                  folderName,
		PackageFlags:                packageFlags,
		NameOffset:                  nameOffset,
		GatherableTextDataCount:     gatherableTextDataCount,
		GatherableTextDataOffset:    gatherableTextDataOffset,
		ExportOffset:                exportOffset,
		ImportOffset:                importOffset,
		DependsOffset:               dependsOffset,
		StringAssetReferencesCount:  stringAssetReferencesCount,
		StringAssetReferencesOffset: stringAssetReferencesOffset,
		SearchableNamesOffset:       searchableNamesOffset,
		ThumbnailTableOffset:        thumbnailTableOffset,
		GUID:                        guid,
		Generations:                 generations,
		SavedByEngineVersion:        savedByEngineVersion,
		CompatibleWithEngineVersion: compatibleWithEngineVersion,
		CompressionFlags:            compressionFlags,
		CompressedChunks:            compressedChunks,
		PackageSource:               packageSource,
		AdditionalPackagesToCook:    additionalPackagesToCook,
		AssetRegistryDataOffset:     assetRegistryDataOffset,
		BulkDataStartOffset:         bulkDataStartOffset,
		WorldTileInfoDataOffset:     worldTileInfoDataOffset,
		ChunkIds:                    chunkIds,
		PreloadDependencyCount:      preloadDependencyCount,
		PreloadDependencyOffset:     preloadDependencyOffset,
		Names:                       names,
		Imports:                     imports,
		Exports:                     exports,
	}
}

func (record *FPakEntry) ReadUExp(pak *PakFile, parser *PakParser, uAsset *FPackageFileSummary) map[*FObjectExport][]*FPropertyTag {
	// Skip UE4 pak header
	// TODO Find out what's in the pak header
	headerSize := int64(pak.Footer.HeaderSize())

	exports := make(map[*FObjectExport][]*FPropertyTag)

	for _, export := range uAsset.Exports {
		log.Debugf("Reading export: %x", headerSize+int64(record.FileOffset)+(export.SerialOffset-int64(uAsset.TotalHeaderSize)))
		parser.Seek(headerSize+int64(record.FileOffset)+(export.SerialOffset-int64(uAsset.TotalHeaderSize)), 0)

		properties := make([]*FPropertyTag, 0)

		for {
			property := parser.ReadFPropertyTag(uAsset, true, 0)

			if property == nil {
				break
			}

			properties = append(properties, property)
		}

		/*
			if len(exportData[offset:]) > 4 {
				fmt.Println()
				spew.Dump(export)
				fmt.Printf("Remaining: %d\n", len(exportData[offset:]))

				if len(exportData[offset:]) < 10000 {
					fmt.Println(utils.HexDump(exportData[offset:]))
				}

				fmt.Println()
			}
		*/

		exports[export] = properties
	}

	// fmt.Println("UEXP LEFTOVERS:", len(fileData[globalOffset:]))
	// fmt.Println(utils.HexDump(fileData[globalOffset:]))

	return exports
}

func (parser *PakParser) ReadFPropertyTag(uAsset *FPackageFileSummary, readData bool, depth int) *FPropertyTag {
	name := parser.ReadFName(uAsset.Names)

	if strings.Trim(name, "\x00") == "None" {
		return nil
	}

	propertyType := parser.ReadFName(uAsset.Names)
	size := parser.ReadInt32()
	arrayIndex := parser.ReadInt32()

	log.Tracef("%sReading Property %s (%s)[%d]", d(depth), strings.Trim(name, "\x00"), strings.Trim(propertyType, "\x00"), size)

	var tagData interface{}

	switch strings.Trim(propertyType, "\x00") {
	case "StructProperty":
		tagData = &StructProperty{
			Type: parser.ReadFName(uAsset.Names),
			Guid: parser.ReadFGuid(),
		}

		log.Tracef("%sStructProperty Type: %s", d(depth), tagData.(*StructProperty).Type)
		break
	case "BoolProperty":
		tagData = parser.Read(1)[0] != 0
		break
	case "EnumProperty":
		fallthrough
	case "ByteProperty":
		fallthrough
	case "SetProperty":
		fallthrough
	case "ArrayProperty":
		tagData = parser.ReadFName(uAsset.Names)
		break
	case "MapProperty":
		tagData = &MapProperty{
			KeyType:   parser.ReadFName(uAsset.Names),
			ValueType: parser.ReadFName(uAsset.Names),
		}
		break
	}

	hasGuid := parser.Read(1)[0] != 0

	var propertyGuid *FGuid

	if hasGuid {
		propertyGuid = parser.ReadFGuid()
	}

	var tag interface{}

	if readData && size > 0 {
		parser.Preload(size)
		parser.TrackRead()
		tag = parser.ReadTag(size, uAsset, propertyType, tagData, &name, depth)

		if parser.tracker.bytesRead != size {
			log.Warningf("%sProperty not read correctly: %d read out of %d", d(depth), parser.tracker.bytesRead, size)

			if parser.tracker.bytesRead > size {
				log.Fatalf("More bytes read than available!")
			} else {
				parser.Read(size - parser.tracker.bytesRead)
			}
		}

		parser.UnTrackRead()
	}

	return &FPropertyTag{
		Name:         name,
		PropertyType: propertyType,
		TagData:      tagData,
		Size:         size,
		ArrayIndex:   arrayIndex,
		PropertyGuid: propertyGuid,
		Tag:          tag,
	}
}

func (parser *PakParser) ReadTag(size int32, uAsset *FPackageFileSummary, propertyType string, tagData interface{}, name *string, depth int) interface{} {
	var tag interface{}
	switch strings.Trim(propertyType, "\x00") {
	case "FloatProperty":
		tag = parser.ReadFloat32()
		break
	case "ArrayProperty":
		arrayTypes := strings.Trim(tagData.(string), "\x00")
		valueCount := parser.ReadInt32()

		var innerTagData *FPropertyTag

		if arrayTypes == "StructProperty" {
			innerTagData = parser.ReadFPropertyTag(uAsset, false, depth+1)
		}

		values := make([]interface{}, valueCount)
		for i := int32(0); i < valueCount; i++ {
			switch arrayTypes {
			case "SoftObjectProperty":
				values[i] = &FSoftObjectPath{
					AssetPathName: parser.ReadFName(uAsset.Names),
					SubPath:       parser.ReadString(),
				}
				break
			case "StructProperty":
				log.Tracef("%sReading Array StructProperty: %s", d(depth), strings.Trim(innerTagData.TagData.(*StructProperty).Type, "\x00"))
				values[i] = &ArrayStructProperty{
					InnerTagData: innerTagData,
					Properties:   parser.ReadTag(-1, uAsset, arrayTypes, innerTagData.TagData, nil, depth+1),
				}
				break
			case "ObjectProperty":
				values[i] = parser.ReadFPackageIndex(uAsset.Imports, uAsset.Exports)
				break
			case "BoolProperty":
				values[i] = parser.Read(1)[0] != 0
				break
			case "ByteProperty":
				if (size-4)/valueCount == 1 {
					values[i] = parser.Read(1)[0]
				} else {
					values[i] = parser.ReadFName(uAsset.Names)
				}
				break
			case "NameProperty":
				fallthrough
			case "EnumProperty":
				values[i] = parser.ReadFName(uAsset.Names)
				break
			case "IntProperty":
				values[i] = parser.ReadInt32()
				break
			case "FloatProperty":
				values[i] = parser.ReadFloat32()
				break
			case "TextProperty":
				values[i] = parser.ReadFText()
				break
			case "StrProperty":
				values[i] = parser.ReadString()
				break
			case "DelegateProperty":
				values[i] = &FScriptDelegate{
					Object: parser.ReadInt32(),
					Name:   parser.ReadFName(uAsset.Names),
				}
				break
			default:
				panic("unknown array type: " + arrayTypes)
			}
		}

		tag = values

		if valueCount > 0 && arrayTypes == "StructProperty" && values[0].(*ArrayStructProperty).Properties == nil {
			if size > 0 {
				// Struct data was not processed
				parser.Read(innerTagData.Size)
			}
		}

		break
	case "StructProperty":
		if tagData == nil {
			log.Trace("%sReading Generic StructProperty", d(depth))
		} else {
			log.Tracef("%sReading StructProperty: %s", d(depth), strings.Trim(tagData.(*StructProperty).Type, "\x00"))

			if structData, ok := tagData.(*StructProperty); ok {
				switch strings.Trim(structData.Type, "\x00") {
				case "Guid":
					fallthrough
				case "VectorMaterialInput":
					fallthrough
				case "ExpressionInput":
					fallthrough
				case "LinearColor":
					fallthrough
				case "ScalarMaterialInput":
					fallthrough
				case "Vector":
					fallthrough
				case "Rotator":
					fallthrough
				case "IntPoint":
					fallthrough
				case "RichCurveKey":
					fallthrough
				case "Vector2D":
					fallthrough
				case "ColorMaterialInput":
					fallthrough
				case "Color":
					fallthrough
				case "Quat":
					fallthrough
				case "Box":
					fallthrough
				case "PerPlatformFloat":
					fallthrough
				case "SkeletalMeshSamplingLODBuiltData":
					fallthrough
				case "PointerToUberGraphFrame":
					fallthrough
				case "MovieSceneFrameRange":
					fallthrough
				case "FrameNumber":
					fallthrough
				case "MovieSceneSegmentIdentifier":
					fallthrough
				case "MovieSceneSequenceID":
					fallthrough
				case "MovieSceneTrackIdentifier":
					fallthrough
				case "MovieSceneEvaluationKey":
					fallthrough
				case "Box2D":
					fallthrough
				case "Vector4":
					fallthrough
				case "FontData":
					fallthrough
				case "FontCharacter":
					fallthrough
				case "MaterialAttributesInput":
					fallthrough
				case "MovieSceneByteChannel":
					fallthrough
				case "MovieSceneEventParameters":
					fallthrough
				case "SoftClassPath":
					fallthrough
				case "MovieSceneParticleChannel":
					fallthrough
				case "InventoryItem":
					fallthrough
				case "SmartName":
					fallthrough
				case "PerPlatformInt":
					fallthrough
				case "MovieSceneFloatValue":
					fallthrough
				case "MovieSceneSegment":
					fallthrough
				case "SectionEvaluationDataTree":
					fallthrough
				case "MovieSceneEvalTemplatePtr":
					fallthrough
				case "MovieSceneTrackImplementationPtr":
					// TODO Read types correctly
					log.Debugf("%sUnread StructProperty Type [%d]: %s", d(depth), size, strings.Trim(structData.Type, "\x00"))
					// fmt.Println(utils.HexDump(data[offset:]))
					if size > 0 {
						parser.Read(size)
					}
					return tag
				default:
					// All others are fine
					break
				}
			}
		}

		properties := make([]*FPropertyTag, 0)

		for {
			property := parser.ReadFPropertyTag(uAsset, true, depth+1)

			if property == nil {
				break
			}

			properties = append(properties, property)
		}

		tag = properties
		break
	case "IntProperty":
		tag = parser.ReadInt32()
		break
	case "Int8Property":
		tag = int8(parser.Read(1)[0])
		break
	case "ObjectProperty":
		tag = parser.ReadFPackageIndex(uAsset.Imports, uAsset.Exports)
		break
	case "TextProperty":
		tag = parser.ReadFText()
		break
	case "BoolProperty":
		// No extra data
		break
	case "NameProperty":
		tag = parser.ReadFName(uAsset.Names)
		break
	case "StrProperty":
		tag = parser.ReadString()
		break
	case "UInt32Property":
		tag = parser.ReadUint32()
		break
	case "UInt64Property":
		tag = parser.ReadUint64()
		break
	case "InterfaceProperty":
		tag = &UInterfaceProperty{
			InterfaceNumber: parser.ReadUint32(),
		}
		break
	case "ByteProperty":
		if size == 4 || size == -4 {
			tag = parser.ReadUint32()
		} else if size >= 8 {
			tag = parser.ReadFName(uAsset.Names)
		} else {
			tag = parser.Read(1)[0]
		}
		break
	case "SoftObjectProperty":
		tag = &FSoftObjectPath{
			AssetPathName: parser.ReadFName(uAsset.Names),
			SubPath:       parser.ReadString(),
		}
		break
	case "EnumProperty":
		if size == 8 {
			tag = parser.ReadFName(uAsset.Names)
		} else if size == 0 {
			break
		} else {
			panic("unknown state!")
		}
		break
	case "MapProperty":
		keyType := tagData.(*MapProperty).KeyType
		valueType := tagData.(*MapProperty).ValueType

		var keyData interface{}
		var valueData interface{}

		realTagData, ok := mapPropertyTypeOverrides[strings.Trim(*name, "\x00")]

		if ok {
			if strings.Trim(keyType, "\x00") != "StructProperty" {
				keyType = realTagData.KeyType
			} else {
				keyData = &StructProperty{
					Type: realTagData.KeyType,
				}
			}

			if strings.Trim(valueType, "\x00") != "StructProperty" {
				valueType = realTagData.ValueType
			} else {
				valueData = &StructProperty{
					Type: realTagData.ValueType,
				}
			}
		}

		if strings.Trim(keyType, "\x00") == "StructProperty" && keyData == nil {
			parser.Read(size)
			log.Warningf("%sSkipping MapProperty [%s] %s -> %s", d(depth), strings.Trim(*name, "\x00"), strings.Trim(keyType, "\x00"), strings.Trim(valueType, "\x00"))
			break
		}

		log.Tracef("%sReading MapProperty [%d]: %s -> %s", d(depth), size, strings.Trim(keyType, "\x00"), strings.Trim(valueType, "\x00"))

		numKeysToRemove := parser.ReadUint32()

		if numKeysToRemove != 0 {
			// TODO Read MapProperty where numKeysToRemove != 0
			parser.Read(size - 4)
			log.Warningf("%sSkipping MapProperty [%s] Remove Key Count: %d", d(depth), strings.Trim(*name, "\x00"), numKeysToRemove)
			break
		}

		num := parser.ReadInt32()

		results := make([]*MapPropertyEntry, num)
		for i := int32(0); i < num; i++ {
			key := parser.ReadTag(-4, uAsset, keyType, keyData, nil, depth+1)

			if key == nil {
				parser.Read(size - 8)
				log.Warningf("%sSkipping MapProperty [%s]: nil key", d(depth), strings.Trim(*name, "\x00"))
				break
			}

			value := parser.ReadTag(-4, uAsset, valueType, valueData, nil, depth+1)

			results[i] = &MapPropertyEntry{
				Key:   key,
				Value: value,
			}
		}

		tag = results
		break
	default:
		log.Debugf("%sUnread Tag Type: %s", d(depth), strings.Trim(propertyType, "\x00"))
		parser.Read(size)
		break
	}

	return tag
}

func d(n int) string {
	return strings.Repeat("  ", n)
}
