package parser

import (
	"encoding/binary"
	"fmt"
	"github.com/Vilsol/ue4pak/utils"
	log "github.com/sirupsen/logrus"
	"math"
	"strings"
)

func ProcessPak(file PakReader, parseFile func(string) bool) ([]*PakEntrySet, *PakFile) {
	pak := Parse(file)

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
			summaries[trimmed[0:strings.Index(trimmed, ".uasset")]] = record.ReadUAsset(pak, file)
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

			exports := record.ReadUExp(pak, file, summary)

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

func Parse(file PakReader) *PakFile {
	// Find magic number
	magicOffset := int64(-44)

	for {
		file.Seek(magicOffset, 2)

		magicArray := make([]byte, 4)
		file.Read(magicArray)

		if magicArray[0] == 0xE1 && magicArray[1] == 0x12 && magicArray[2] == 0x6F && magicArray[3] == 0x5A {
			break
		}

		magicOffset -= 1

		if magicOffset < -1024 {
			log.Fatal("Could not find magic bytes in pak!")
		}
	}

	// Seek and read the footer of the file
	file.Seek(magicOffset, 2)
	footer := make([]byte, magicOffset*-1)
	file.Read(footer)

	pakFooter := &FPakInfo{
		Magic:         binary.LittleEndian.Uint32(footer[0:4]),
		Version:       binary.LittleEndian.Uint32(footer[4:8]),
		IndexOffset:   binary.LittleEndian.Uint64(footer[8:16]),
		IndexSize:     binary.LittleEndian.Uint64(footer[16:24]),
		IndexSHA1Hash: footer[24:44],
	}

	// Seek and read the index of the file
	file.Seek(int64(pakFooter.IndexOffset), 0)
	index := make([]byte, pakFooter.IndexSize)
	file.Read(index)

	offset := uint32(0)

	mountPoint, tempOffset := ReadString(index[offset:])
	offset += tempOffset

	recordCount := binary.LittleEndian.Uint32(index[offset:])
	offset += 4

	pakIndex := &FPakIndex{
		MountPoint: mountPoint,
		Records:    make([]*FPakEntry, recordCount),
	}

	for i := 0; i < len(pakIndex.Records); i++ {
		pakIndex.Records[i] = &FPakEntry{}

		var tempOffset uint32
		pakIndex.Records[i].FileName, tempOffset = ReadString(index[offset:])
		offset += tempOffset

		pakIndex.Records[i].FileOffset = binary.LittleEndian.Uint64(index[offset:])
		offset += 8

		pakIndex.Records[i].FileSize = binary.LittleEndian.Uint64(index[offset:])
		offset += 8

		pakIndex.Records[i].UncompressedSize = binary.LittleEndian.Uint64(index[offset:])
		offset += 8

		if pakFooter.Version >= 8 {
			pakIndex.Records[i].CompressionMethod = uint32(index[offset])
			offset += 1
		} else {
			pakIndex.Records[i].CompressionMethod = binary.LittleEndian.Uint32(index[offset:])
			offset += 4
		}

		if pakFooter.Version <= 1 {
			pakIndex.Records[i].Timestamp = binary.LittleEndian.Uint64(index[offset:])
			offset += 8
		}

		pakIndex.Records[i].DataSHA1Hash = index[offset : offset+20]
		offset += 20

		if pakFooter.Version >= 3 {
			if pakIndex.Records[i].CompressionMethod != 0 {
				blockCount := binary.LittleEndian.Uint32(index[offset:])
				offset += 4

				pakIndex.Records[i].CompressionBlocks = make([]*FPakCompressedBlock, blockCount)

				for j := 0; j < len(pakIndex.Records[i].CompressionBlocks); j++ {
					pakIndex.Records[i].CompressionBlocks[j] = &FPakCompressedBlock{
						StartOffset: binary.LittleEndian.Uint64(index[offset:]),
						EndOffset:   binary.LittleEndian.Uint64(index[offset+8:]),
					}
					offset += 16
				}
			}

			pakIndex.Records[i].IsEncrypted = index[offset] > 0
			offset += 1

			pakIndex.Records[i].CompressionBlockSize = binary.LittleEndian.Uint32(index[offset:])
			offset += 4
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

func (record *FPakEntry) ReadUAsset(pak *PakFile, file PakReader) *FPackageFileSummary {
	// Skip UE4 pak header
	// TODO Find out what's in the pak header
	headerSize := int64(pak.Footer.HeaderSize())

	file.Seek(headerSize+int64(record.FileOffset), 0)
	fileData := make([]byte, record.FileSize)
	file.Read(fileData)

	offset := uint32(0)

	tag := utils.Int32(fileData[offset:])
	offset += 4

	legacyFileVersion := utils.Int32(fileData[offset:])
	offset += 4

	legacyUE3Version := utils.Int32(fileData[offset:])
	offset += 4

	fileVersionUE4 := utils.Int32(fileData[offset:])
	offset += 4

	fileVersionLicenseeUE4 := utils.Int32(fileData[offset:])
	offset += 4

	// TODO custom_version_container: Vec<FCustomVersion>
	offset += 4

	totalHeaderSize := utils.Int32(fileData[offset:])
	offset += 4

	folderName, tempOffset := ReadString(fileData[offset:])
	offset += tempOffset

	packageFlags := binary.LittleEndian.Uint32(fileData[offset:])
	offset += 4

	nameCount := binary.LittleEndian.Uint32(fileData[offset:])
	offset += 4

	nameOffset := utils.Int32(fileData[offset:])
	offset += 4

	gatherableTextDataCount := utils.Int32(fileData[offset:])
	offset += 4

	gatherableTextDataOffset := utils.Int32(fileData[offset:])
	offset += 4

	exportCount := binary.LittleEndian.Uint32(fileData[offset:])
	offset += 4

	exportOffset := utils.Int32(fileData[offset:])
	offset += 4

	importCount := binary.LittleEndian.Uint32(fileData[offset:])
	offset += 4

	importOffset := utils.Int32(fileData[offset:])
	offset += 4

	dependsOffset := utils.Int32(fileData[offset:])
	offset += 4

	stringAssetReferencesCount := utils.Int32(fileData[offset:])
	offset += 4

	stringAssetReferencesOffset := utils.Int32(fileData[offset:])
	offset += 4

	searchableNamesOffset := utils.Int32(fileData[offset:])
	offset += 4

	thumbnailTableOffset := utils.Int32(fileData[offset:])
	offset += 4

	guid := ReadFGuid(fileData[offset:])
	offset += 16

	generationCount := binary.LittleEndian.Uint32(fileData[offset:])
	offset += 4

	generations := make([]*FGenerationInfo, generationCount)
	for i := uint32(0); i < generationCount; i++ {
		generation, tempOffset := ReadFGenerationInfo(fileData[offset:])
		generations[i] = generation
		offset += tempOffset
	}

	savedByEngineVersion, tempOffset := ReadFEngineVersion(fileData[offset:])
	offset += tempOffset

	compatibleWithEngineVersion, tempOffset := ReadFEngineVersion(fileData[offset:])
	offset += tempOffset

	compressionFlags := binary.LittleEndian.Uint32(fileData[offset:])
	offset += 4

	compressedChunkCount := binary.LittleEndian.Uint32(fileData[offset:])
	offset += 4

	compressedChunks := make([]*FCompressedChunk, compressedChunkCount)
	for i := uint32(0); i < compressedChunkCount; i++ {
		compressedChunk, tempOffset := ReadFCompressedChunk(fileData[offset:])
		compressedChunks[i] = compressedChunk
		offset += tempOffset
	}

	packageSource := binary.LittleEndian.Uint32(fileData[offset:])
	offset += 4

	additionalPackageCount := binary.LittleEndian.Uint32(fileData[offset:])
	offset += 4

	additionalPackagesToCook := make([]string, additionalPackageCount)
	for i := uint32(0); i < additionalPackageCount; i++ {
		var tempOffset uint32
		additionalPackagesToCook[i], tempOffset = ReadString(fileData[offset:])
		offset += tempOffset
	}

	assetRegistryDataOffset := utils.Int32(fileData[offset:])
	offset += 4

	bulkDataStartOffset := utils.Int32(fileData[offset:])
	offset += 4

	worldTileInfoDataOffset := utils.Int32(fileData[offset:])
	offset += 4

	chunkCount := binary.LittleEndian.Uint32(fileData[offset:])
	offset += 4

	chunkIds := make([]int32, chunkCount)
	for i := uint32(0); i < chunkCount; i++ {
		chunkIds[i] = utils.Int32(fileData[offset:])
		offset += 4
	}

	// TODO unknown bytes
	offset += 4

	preloadDependencyCount := utils.Int32(fileData[offset:])
	offset += 4

	preloadDependencyOffset := utils.Int32(fileData[offset:])
	offset += 4

	names := make([]*FNameEntrySerialized, nameCount)
	for i := uint32(0); i < nameCount; i++ {
		name, tempOffset := ReadString(fileData[offset:])
		offset += tempOffset

		// fmt.Println(i, name)

		names[i] = &FNameEntrySerialized{
			Name:                  name,
			NonCasePreservingHash: binary.LittleEndian.Uint16(fileData[offset:]),
			CasePreservingHash:    binary.LittleEndian.Uint16(fileData[offset+2:]),
		}

		offset += 4
	}

	imports := make([]*FObjectImport, importCount)
	for i := uint32(0); i < importCount; i++ {
		classPackage, tempOffset := ReadFName(fileData[offset:], names)
		offset += tempOffset

		className, tempOffset := ReadFName(fileData[offset:], names)
		offset += tempOffset

		outerIndex := utils.Int32(fileData[offset:])
		offset += 4

		objectName, tempOffset := ReadFName(fileData[offset:], names)
		offset += tempOffset

		// fmt.Println("IMP", objectName)

		imports[i] = &FObjectImport{
			ClassPackage: classPackage,
			ClassName:    className,
			OuterIndex:   outerIndex,
			ObjectName:   objectName,
		}
	}

	exports := make([]*FObjectExport, exportCount)
	for i := uint32(0); i < exportCount; i++ {
		classIndex := ReadFPackageIndex(fileData[offset:], imports, exports)
		offset += 4

		superIndex := ReadFPackageIndex(fileData[offset:], imports, exports)
		offset += 4

		templateIndex := ReadFPackageIndex(fileData[offset:], imports, exports)
		offset += 4

		outerIndex := ReadFPackageIndex(fileData[offset:], imports, exports)
		offset += 4

		objectName, tempOffset := ReadFName(fileData[offset:], names)
		offset += tempOffset

		save := binary.LittleEndian.Uint32(fileData[offset:])
		offset += 4

		serialSize := utils.Int64(fileData[offset:])
		offset += 8

		serialOffset := utils.Int64(fileData[offset:])
		offset += 8

		forcedExport := utils.Int32(fileData[offset:]) != 0
		offset += 4

		notForClient := utils.Int32(fileData[offset:]) != 0
		offset += 4

		notForServer := utils.Int32(fileData[offset:]) != 0
		offset += 4

		packageGuid := ReadFGuid(fileData[offset:])
		offset += 16

		packageFlags := binary.LittleEndian.Uint32(fileData[offset:])
		offset += 4

		notAlwaysLoadedForEditorGame := utils.Int32(fileData[offset:]) != 0
		offset += 4

		isAsset := utils.Int32(fileData[offset:]) != 0
		offset += 4

		firstExportDependency := utils.Int32(fileData[offset:])
		offset += 4

		serializationBeforeSerializationDependencies := utils.Int32(fileData[offset:]) != 0
		offset += 4

		createBeforeSerializationDependencies := utils.Int32(fileData[offset:]) != 0
		offset += 4

		serializationBeforeCreateDependencies := utils.Int32(fileData[offset:]) != 0
		offset += 4

		createBeforeCreateDependencies := utils.Int32(fileData[offset:]) != 0
		offset += 4

		// fmt.Println("EXP", objectName)

		exports[i] = &FObjectExport{
			ClassIndex:                   classIndex,
			SuperIndex:                   superIndex,
			TemplateIndex:                templateIndex,
			OuterIndex:                   outerIndex,
			ObjectName:                   objectName,
			Save:                         save,
			SerialSize:                   serialSize,
			SerialOffset:                 serialOffset,
			ForcedExport:                 forcedExport,
			NotForClient:                 notForClient,
			NotForServer:                 notForServer,
			PackageGuid:                  packageGuid,
			PackageFlags:                 packageFlags,
			NotAlwaysLoadedForEditorGame: notAlwaysLoadedForEditorGame,
			IsAsset:                      isAsset,
			FirstExportDependency:        firstExportDependency,
			SerializationBeforeSerializationDependencies: serializationBeforeSerializationDependencies,
			CreateBeforeSerializationDependencies:        createBeforeSerializationDependencies,
			SerializationBeforeCreateDependencies:        serializationBeforeCreateDependencies,
			CreateBeforeCreateDependencies:               createBeforeCreateDependencies,
		}
	}

	for _, objectImport := range imports {
		objectImport.OuterPackage = ReadFPackageIndexInt(objectImport.OuterIndex, imports, exports)
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

func (record *FPakEntry) ReadUExp(pak *PakFile, file PakReader, uAsset *FPackageFileSummary) map[*FObjectExport][]*FPropertyTag {
	// Skip UE4 pak header
	// TODO Find out what's in the pak header
	headerSize := int64(pak.Footer.HeaderSize())

	file.Seek(headerSize+int64(record.FileOffset), 0)
	fileData := make([]byte, record.FileSize)
	file.Read(fileData)

	exports := make(map[*FObjectExport][]*FPropertyTag)

	// globalOffset := int64(0)

	for _, export := range uAsset.Exports {
		headerOffset := export.SerialOffset - int64(uAsset.TotalHeaderSize)

		exportData := fileData[headerOffset : headerOffset+export.SerialSize]

		// globalOffset = headerOffset + export.SerialSize

		offset := uint32(0)

		properties := make([]*FPropertyTag, 0)

		for {
			property, tempOffset := ReadFPropertyTag(exportData[offset:], uAsset.Imports, uAsset.Exports, uAsset.Names, true)
			offset += tempOffset

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

func ReadFName(data []byte, names []*FNameEntrySerialized) (string, uint32) {
	return names[binary.LittleEndian.Uint32(data)].Name, 8
}

func ReadFPackageIndex(data []byte, imports []*FObjectImport, exports []*FObjectExport) *FPackageIndex {
	return ReadFPackageIndexInt(utils.Int32(data), imports, exports)
}

func ReadFPackageIndexInt(index int32, imports []*FObjectImport, exports []*FObjectExport) *FPackageIndex {
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

func ReadFGuid(data []byte) *FGuid {
	return &FGuid{
		A: binary.LittleEndian.Uint32(data),
		B: binary.LittleEndian.Uint32(data[4:]),
		C: binary.LittleEndian.Uint32(data[8:]),
		D: binary.LittleEndian.Uint32(data[12:]),
	}
}

func ReadFPropertyTag(data []byte, imports []*FObjectImport, exports []*FObjectExport, names []*FNameEntrySerialized, readData bool) (*FPropertyTag, uint32) {
	offset := uint32(0)

	name, tempOffset := ReadFName(data[offset:], names)
	offset += tempOffset

	if strings.Trim(name, "\x00") == "None" {
		return nil, offset
	}

	propertyType, tempOffset := ReadFName(data[offset:], names)
	offset += tempOffset

	log.Tracef("Reading Property %s (%s)", strings.Trim(name, "\x00"), strings.Trim(propertyType, "\x00"))

	size := utils.Int32(data[offset:])
	offset += 4

	arrayIndex := utils.Int32(data[offset:])
	offset += 4

	var tagData interface{}

	switch strings.Trim(propertyType, "\x00") {
	case "StructProperty":
		structType, tempOffset := ReadFName(data[offset:], names)
		offset += tempOffset

		log.Tracef("StructProperty Type: %s", structType)

		structGuid := ReadFGuid(data[offset:])
		offset += 16

		tagData = &StructProperty{
			Type: structType,
			Guid: structGuid,
		}
		break
	case "BoolProperty":
		tagData = data[offset] != 0
		offset += 1
		break
	case "EnumProperty":
		fallthrough
	case "ByteProperty":
		fallthrough
	case "SetProperty":
		fallthrough
	case "ArrayProperty":
		tagData, tempOffset = ReadFName(data[offset:], names)
		offset += tempOffset
		break
	case "MapProperty":
		keyType, tempOffset := ReadFName(data[offset:], names)
		offset += tempOffset

		valueType, tempOffset := ReadFName(data[offset:], names)
		offset += tempOffset

		tagData = &MapProperty{
			KeyType:   keyType,
			ValueType: valueType,
		}
		break
	}

	hasGuid := data[offset] != 0
	offset += 1

	var propertyGuid *FGuid

	if hasGuid {
		propertyGuid = ReadFGuid(data[offset:])
		offset += 16
	}

	var tag interface{}

	if readData {
		tag, tempOffset = ReadTag(data[offset:offset+uint32(size)], imports, exports, names, propertyType, tagData)

		if tempOffset != uint32(size) {
			log.Debugf("Property not read to end: %s (%s)", strings.Trim(name, "\x00"), strings.Trim(propertyType, "\x00"))
		}

		offset += uint32(size)
	}

	return &FPropertyTag{
		Name:         name,
		PropertyType: propertyType,
		TagData:      tagData,
		Size:         size,
		ArrayIndex:   arrayIndex,
		PropertyGuid: propertyGuid,
		Tag:          tag,
	}, offset
}

func ReadTag(data []byte, imports []*FObjectImport, exports []*FObjectExport, names []*FNameEntrySerialized, propertyType string, tagData interface{}) (interface{}, uint32) {
	offset := uint32(0)

	var tempOffset uint32
	var tag interface{}
	switch strings.Trim(propertyType, "\x00") {
	case "FloatProperty":
		tag = math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4
		break
	case "ArrayProperty":
		arrayTypes := strings.Trim(tagData.(string), "\x00")
		valueCount := utils.Int32(data[offset:])
		offset += 4

		bufferLength := len(data[offset:])

		var innerTagData *FPropertyTag

		if arrayTypes == "StructProperty" {
			innerTagData, tempOffset = ReadFPropertyTag(data[offset:], imports, exports, names, false)
			offset += tempOffset
		}

		values := make([]interface{}, valueCount)
		for i := int32(0); i < valueCount; i++ {
			switch arrayTypes {
			case "SoftObjectProperty":
				assetPathName, tempOffset := ReadFName(data[offset:], names)
				offset += tempOffset

				subPath, tempOffset := ReadString(data[offset:])
				offset += tempOffset

				values[i] = &FSoftObjectPath{
					AssetPathName: assetPathName,
					SubPath:       subPath,
				}
				break
			case "StructProperty":
				log.Debugf("Reading Array StructProperty: %s", strings.Trim(innerTagData.TagData.(*StructProperty).Type, "\x00"))
				var structProperties interface{}
				structProperties, tempOffset = ReadTag(data[offset:], imports, exports, names, arrayTypes, innerTagData.TagData)
				offset += tempOffset

				values[i] = &ArrayStructProperty{
					InnerTagData: innerTagData,
					Properties:   structProperties,
				}
				break
			case "ObjectProperty":
				values[i] = ReadFPackageIndex(data[offset:], imports, exports)
				offset += 4
				break
			case "BoolProperty":
				values[i] = data[offset] != 0
				offset += 1
				break
			case "ByteProperty":
				if int32(bufferLength)/valueCount == 1 {
					values[i] = uint8(data[offset])
					offset += 1
				} else {
					values[i], tempOffset = ReadFName(data[offset:], names)
					offset += tempOffset
				}
				break
			case "NameProperty":
				fallthrough
			case "EnumProperty":
				values[i], tempOffset = ReadFName(data[offset:], names)
				offset += tempOffset
				break
			case "IntProperty":
				values[i] = utils.Int32(data[offset:])
				offset += 4
				break
			case "FloatProperty":
				values[i] = utils.Float32(data[offset:])
				offset += 4
				break
			case "TextProperty":
				values[i], tempOffset = ReadFText(data[offset:])
				offset += tempOffset
				break
			case "StrProperty":
				values[i], tempOffset = ReadString(data[offset:])
				offset += tempOffset
				break
			case "DelegateProperty":
				object := utils.Int32(data[offset:])
				offset += 4

				name, tempOffset := ReadFName(data[offset:], names)
				offset += tempOffset

				values[i] = &FScriptDelegate{
					Object: object,
					Name:   name,
				}
				break
			default:
				fmt.Println(utils.HexDump(data[offset:]))
				panic("unknown type: " + arrayTypes)
			}
		}

		tag = values

		break
	case "StructProperty":
		if tagData == nil {
			log.Debug("Reading Generic StructProperty")
		} else {
			log.Debugf("Reading StructProperty: %s", strings.Trim(tagData.(*StructProperty).Type, "\x00"))

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
					log.Debugf("Unread StructProperty Type: %s", strings.Trim(structData.Type, "\x00"))
					// fmt.Println(utils.HexDump(data[offset:]))
					offset = uint32(len(data))
					break
				default:
					// All others are fine
					break
				}

				if offset == uint32(len(data)) {
					break
				}
			}
		}

		properties := make([]*FPropertyTag, 0)

		for {
			property, tempOffset := ReadFPropertyTag(data[offset:], imports, exports, names, true)
			offset += tempOffset

			if property == nil {
				break
			}

			properties = append(properties, property)
		}

		tag = properties
		break
	case "IntProperty":
		tag = utils.Int32(data[offset : offset+4])
		offset += 4
		break
	case "Int8Property":
		tag = int8(data[offset])
		offset += 1
		break
	case "ObjectProperty":
		tag = ReadFPackageIndex(data[offset:], imports, exports)
		offset += 4
		break
	case "TextProperty":
		tag, tempOffset = ReadFText(data[offset:])
		offset += tempOffset
		break
	case "BoolProperty":
		// No extra data
		break
	case "NameProperty":
		tag, tempOffset = ReadFName(data[offset:], names)
		offset += tempOffset
		break
	case "StrProperty":
		tag, tempOffset = ReadString(data[offset:])
		offset += tempOffset
		break
	case "UInt32Property":
		tag = binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		break
	case "UInt64Property":
		tag = binary.LittleEndian.Uint64(data[offset:])
		offset += 8
		break
	case "InterfaceProperty":
		tag = &UInterfaceProperty{
			InterfaceNumber: binary.LittleEndian.Uint32(data[offset:]),
		}
		offset += 4
		break
	case "ByteProperty":
		if len(data[offset:]) == 4 {
			tag = binary.LittleEndian.Uint32(data[offset:])
			offset += 4
		} else if len(data[offset:]) >= 8 {
			tag, tempOffset = ReadFName(data[offset:], names)
			offset += tempOffset
		} else {
			tag = data[offset]
			offset += 1
		}
		break
	case "SoftObjectProperty":
		assetPathName, tempOffset := ReadFName(data[offset:], names)
		offset += tempOffset

		subPath, tempOffset := ReadString(data[offset:])
		offset += tempOffset

		tag = &FSoftObjectPath{
			AssetPathName: assetPathName,
			SubPath:       subPath,
		}
		break
	case "EnumProperty":
		if len(data[offset:]) == 8 {
			tag, tempOffset = ReadFName(data[offset:], names)
			offset += tempOffset
		} else if len(data[offset:]) == 0 {
			break
		} else {
			panic("unknown state!")
		}
		break
	case "MapProperty":
		keyType := tagData.(*MapProperty).KeyType
		valueType := tagData.(*MapProperty).ValueType

		if strings.Trim(keyType, "\x00") == "StructProperty" {
			// TODO Read MapProperty where StructProperty is key type
			offset += uint32(len(data)) - offset
			log.Warningf("Skipping MapProperty: %s -> %s", strings.Trim(keyType, "\x00"), strings.Trim(valueType, "\x00"))
			break
		}

		log.Debugf("Reading MapProperty: %s -> %s", strings.Trim(keyType, "\x00"), strings.Trim(valueType, "\x00"))

		numKeysToRemove := binary.LittleEndian.Uint32(data[offset:])
		offset += 4

		if numKeysToRemove != 0 {
			// TODO Read MapProperty where numKeysToRemove != 0
			offset += uint32(len(data)) - offset
			log.Warningf("Skipping MapProperty: Remove Key Count: %d", numKeysToRemove)
			break
		}

		num := utils.Int32(data[offset:])
		offset += 4

		results := make([]*MapPropertyEntry, num)
		for i := int32(0); i < num; i++ {
			key, tempOffset := ReadTag(data[offset:], imports, exports, names, keyType, nil)
			offset += tempOffset

			value, tempOffset := ReadTag(data[offset:], imports, exports, names, valueType, nil)
			offset += tempOffset

			results[i] = &MapPropertyEntry{
				Key:   key,
				Value: value,
			}
		}

		tag = results
		break
	default:
		log.Debugf("Unread Tag Type: %s", strings.Trim(propertyType, "\x00"))

		if offset < uint32(len(data)-1) {
			offset += uint32(len(data)) - offset
		}
		break
	}

	return tag, offset
}

func ReadFEngineVersion(data []byte) (*FEngineVersion, uint32) {
	offset := uint32(0)

	major := binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	minor := binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	patch := binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	changeList := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	branch, tempOffset := ReadString(data[offset:])
	offset += tempOffset

	return &FEngineVersion{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		ChangeList: changeList,
		Branch:     branch,
	}, offset
}

func ReadFGenerationInfo(data []byte) (*FGenerationInfo, uint32) {
	offset := uint32(0)

	exportCount := utils.Int32(data[offset:])
	offset += 4

	nameCount := utils.Int32(data[offset:])
	offset += 4

	return &FGenerationInfo{
		ExportCount: exportCount,
		NameCount:   nameCount,
	}, offset
}

func ReadFCompressedChunk(data []byte) (*FCompressedChunk, uint32) {
	offset := uint32(0)

	uncompressedOffset := utils.Int32(data[offset:])
	offset += 4

	uncompressedSize := utils.Int32(data[offset:])
	offset += 4

	compressedOffset := utils.Int32(data[offset:])
	offset += 4

	compressedSize := utils.Int32(data[offset:])
	offset += 4

	return &FCompressedChunk{
		UncompressedOffset: uncompressedOffset,
		UncompressedSize:   uncompressedSize,
		CompressedOffset:   compressedOffset,
		CompressedSize:     compressedSize,
	}, offset
}

func ReadFText(data []byte) (*FText, uint32) {
	offset := uint32(0)

	flags := binary.LittleEndian.Uint32(data)
	offset += 4

	historyType := int8(data[offset])
	offset += 1

	if historyType != 0 {
		return &FText{
			Flags:       flags,
			HistoryType: historyType,
		}, offset
	}

	namespace, tempOffset := ReadString(data[offset:])
	offset += tempOffset

	key, tempOffset := ReadString(data[offset:])
	offset += tempOffset

	sourceString, tempOffset := ReadString(data[offset:])
	offset += tempOffset

	return &FText{
		Flags:        flags,
		HistoryType:  historyType,
		Namespace:    namespace,
		Key:          key,
		SourceString: sourceString,
	}, offset
}

func ReadString(data []byte) (string, uint32) {
	offset := uint32(0)

	stringLength := utils.Int32(data[offset:])
	offset += 4

	if stringLength > int32(len(data[offset:])) {
		fmt.Println(utils.HexDump(data))
		panic("string length longer than data buffer")
	}

	if stringLength == 0 {
		return "", 4
	}

	if stringLength < 0 {
		stringLength = (stringLength * -1) * 2
		return utils.DecodeUtf16(data[offset : offset+uint32(stringLength)]), offset + uint32(stringLength)
	}

	return string(data[offset : offset+uint32(stringLength)]), offset + uint32(stringLength)
}
