package parser

import (
	"encoding/binary"
	"github.com/Vilsol/ue4pak/utils"
	"math"
	"os"
	"strings"
)

func Parse(file *os.File) *PakFile {
	// Seek and read the footer of the file
	file.Seek(-44, 2)
	footer := make([]byte, 44)
	file.Read(footer)

	pakFooter := &FPakInfo{
		Magic:         binary.LittleEndian.Uint32(footer[0:4]),
		Version:       binary.LittleEndian.Uint32(footer[4:8]),
		IndexOffset:   binary.LittleEndian.Uint64(footer[8:16]),
		IndexSize:     binary.LittleEndian.Uint64(footer[16:24]),
		IndexSHA1Hash: footer[24:],
	}

	// Seek and read the index of the file
	file.Seek(int64(pakFooter.IndexOffset), 0)
	index := make([]byte, pakFooter.IndexSize)
	file.Read(index)

	mountPointLength := binary.LittleEndian.Uint32(index[0:4])
	pakIndex := &FPakIndex{
		MountPoint: string(index[4:mountPointLength]),
		Records:    make([]*FPakEntry, binary.LittleEndian.Uint32(index[4+mountPointLength:])),
	}

	offset := 4 + mountPointLength + 4
	for i := 0; i < len(pakIndex.Records); i++ {
		pakIndex.Records[i] = &FPakEntry{}

		fileNameLength := binary.LittleEndian.Uint32(index[offset:])
		offset += 4

		pakIndex.Records[i].FileName = string(index[offset : offset+fileNameLength])
		offset += fileNameLength

		pakIndex.Records[i].FileOffset = binary.LittleEndian.Uint64(index[offset:])
		offset += 8

		pakIndex.Records[i].FileSize = binary.LittleEndian.Uint64(index[offset:])
		offset += 8

		pakIndex.Records[i].UncompressedSize = binary.LittleEndian.Uint64(index[offset:])
		offset += 8

		pakIndex.Records[i].CompressionMethod = binary.LittleEndian.Uint32(index[offset:])
		offset += 4

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

func (record *FPakEntry) ReadUAsset(file *os.File) *FPackageFileSummary {
	// Skip UE4 pak header
	// TODO Find out what's in the pak header
	headerSize := int64(53)

	file.Seek(headerSize+int64(record.FileOffset), 0)
	fileData := make([]byte, record.FileSize)
	file.Read(fileData)

	offset := uint32(0)

	tag := utils.Int32(fileData[offset:])
	offset += 4

	legacyFileVersion := utils.Int32(fileData[offset:])
	offset += 4

	legacyyUE3Version := utils.Int32(fileData[offset:])
	offset += 4

	fileVersionU34 := utils.Int32(fileData[offset:])
	offset += 4

	fileVersionLicenseeUE4 := utils.Int32(fileData[offset:])
	offset += 4

	// TODO custom_version_container: Vec<FCustomVersion>
	offset += 4

	totalHeaderSize := utils.Int32(fileData[offset:])
	offset += 4

	folderNameLength := binary.LittleEndian.Uint32(fileData[offset:])
	offset += 4

	folderName := string(fileData[offset : offset+folderNameLength])
	offset += folderNameLength

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
		packageLength := binary.LittleEndian.Uint32(fileData[offset:])
		offset += 4

		additionalPackagesToCook[i] = string(fileData[offset : offset+packageLength])
		offset += packageLength
	}

	assetRegistryDataOffset := utils.Int32(fileData[offset:])
	offset += 4

	buldDataStartOffset := utils.Int32(fileData[offset:])
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
		strLen := binary.LittleEndian.Uint32(fileData[offset:])
		offset += 4

		names[i] = &FNameEntrySerialized{
			Name:                  string(fileData[offset : offset+strLen]),
			NonCasePreservingHash: binary.LittleEndian.Uint16(fileData[offset+strLen:]),
			CasePreservingHash:    binary.LittleEndian.Uint16(fileData[offset+strLen:]),
		}

		offset += strLen
		offset += 4
	}

	imports := make([]*FObjectImport, importCount)
	for i := uint32(0); i < importCount; i++ {
		classPackage, tempOffset := ReadFName(fileData[offset:], names)
		offset += tempOffset

		className, tempOffset := ReadFName(fileData[offset:], names)
		offset += tempOffset

		outerIndex := binary.LittleEndian.Uint32(fileData[offset:])
		offset += 4

		objectName, tempOffset := ReadFName(fileData[offset:], names)
		offset += tempOffset

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

	// TODO Bunch of unknown bytes at the end

	return &FPackageFileSummary{
		Tag:                         tag,
		LegacyFileVersion:           legacyFileVersion,
		LegacyUE3Version:            legacyyUE3Version,
		FileVersionU34:              fileVersionU34,
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
		BuldDataStartOffset:         buldDataStartOffset,
		WorldTileInfoDataOffset:     worldTileInfoDataOffset,
		ChunkIds:                    chunkIds,
		PreloadDependencyCount:      preloadDependencyCount,
		PreloadDependencyOffset:     preloadDependencyOffset,
		Names:                       names,
		Imports:                     imports,
		Exports:                     exports,
	}
}

func (record *FPakEntry) ReadUExp(file *os.File, uAsset *FPackageFileSummary) map[*FObjectExport][]*FPropertyTag {
	// Skip UE4 pak header
	// TODO Find out what's in the pak header
	headerSize := int64(53)

	file.Seek(headerSize+int64(record.FileOffset), 0)
	fileData := make([]byte, record.FileSize)
	file.Read(fileData)

	exports := make(map[*FObjectExport][]*FPropertyTag)

	for _, export := range uAsset.Exports {
		headerOffset := export.SerialOffset - int64(uAsset.TotalHeaderSize)

		exportData := fileData[headerOffset : headerOffset+export.SerialSize]

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
				fmt.Printf("%#v\n", export.ClassIndex.Reference)
				fmt.Printf("%#v\n", export.TemplateIndex.Reference)
				fmt.Printf("%#v\n", export.SuperIndex.Reference)
				fmt.Printf("%#v\n", export.OuterIndex.Reference)
				fmt.Printf("Remaining: %d\n", len(exportData[offset:]))

				if len(exportData[offset:]) < 10000 {
					fmt.Println(utils.HexDump(exportData[offset:]))
				}

				fmt.Println()
			}
		*/

		exports[export] = properties
	}

	return exports
}

func ReadFName(data []byte, names []*FNameEntrySerialized) (string, uint32) {
	return names[binary.LittleEndian.Uint32(data)].Name, 8
}

func ReadFPackageIndex(data []byte, imports []*FObjectImport, exports []*FObjectExport) *FPackageIndex {
	index := utils.Int32(data)

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

	if index < int32(len(exports)) {
		return &FPackageIndex{
			Index:     index,
			Reference: exports[index],
		}
	}

	return &FPackageIndex{
		Index:     index,
		Reference: nil,
	}
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

	size := utils.Int32(data[offset:])
	offset += 4

	arrayIndex := utils.Int32(data[offset:])
	offset += 4

	var tagData interface{}

	switch strings.Trim(propertyType, "\x00") {
	case "StructProperty":
		structType, tempOffset := ReadFName(data[offset:], names)
		offset += tempOffset

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
		tagData, tempOffset = ReadFName(data[offset:], names)
		offset += tempOffset
		break
	case "ByteProperty":
		tagData, tempOffset = ReadFName(data[offset:], names)
		offset += tempOffset
		break
	case "ArrayProperty":
		tagData, tempOffset = ReadFName(data[offset:], names)
		offset += tempOffset
		break
	case "MapProperty":
		panic("Reading MapProperty not implemented") // TODO Read MapProperty
		break
	case "SetProperty":
		tagData, tempOffset = ReadFName(data[offset:], names)
		offset += tempOffset
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
		offset += tempOffset
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

				subPathLength := binary.LittleEndian.Uint32(data[offset:])
				offset += 4

				subPath := string(data[offset : offset+subPathLength])
				offset += subPathLength

				values[i] = &SoftObjectProperty{
					AssetPathName: assetPathName,
					SubPath:       subPath,
				}
				break
			case "StructProperty":
				values[i], tempOffset = ReadTag(data[offset:], imports, exports, names, arrayTypes, innerTagData.TagData)
				offset += tempOffset
				break
			case "ObjectProperty":
				values[i] = ReadFPackageIndex(data[offset:], imports, exports)
				offset += 4
				break
			case "BoolProperty":
				values[i] = data[offset] != 0
				offset += 1
				break
			default:
				panic("unknown type:" + arrayTypes)
			}
		}

		tag = values

		break
	case "StructProperty":
		if tagData != nil {
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
					// TODO Read types correctly
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
	case "ObjectProperty":
		tag = ReadFPackageIndex(data[offset:], imports, exports)
		offset += 4
		break
	case "TextProperty":
		flags := binary.LittleEndian.Uint32(data[offset:])
		offset += 4

		historyType := int8(data[offset])
		offset += 1

		if historyType == -1 {
			tag = &TextProperty{
				Flags:       flags,
				HistoryType: historyType,
			}
			break
		}

		namespaceLength := binary.LittleEndian.Uint32(data[offset:])
		offset += 4

		namespace := string(data[offset : offset+namespaceLength])
		offset += namespaceLength

		keyLength := binary.LittleEndian.Uint32(data[offset:])
		offset += 4

		key := string(data[offset : offset+keyLength])
		offset += keyLength

		sourceStringLength := binary.LittleEndian.Uint32(data[offset:])
		offset += 4

		sourceString := string(data[offset : offset+sourceStringLength])
		offset += sourceStringLength

		tag = &TextProperty{
			Flags:        flags,
			HistoryType:  historyType,
			Namespace:    namespace,
			Key:          key,
			SourceString: sourceString,
		}
		break
	case "BoolProperty":
		// No extra data
		break
	case "NameProperty":
		tag, tempOffset = ReadFName(data[offset:], names)
		offset += tempOffset
		break
	case "UInt64Property":
		tag = binary.LittleEndian.Uint64(data[offset:])
		offset += 8
		break
	case "ByteProperty":
		if len(data[offset:]) == 4 {
			tag = binary.LittleEndian.Uint32(data[offset:])
			offset += 4
		} else {
			tag, tempOffset = ReadFName(data[offset:], names)
			offset += tempOffset
		}
		break
	case "SoftObjectProperty":
		assetPathName, tempOffset := ReadFName(data[offset:], names)
		offset += tempOffset

		subPathLength := binary.LittleEndian.Uint32(data[offset:])
		offset += 4

		subPath := string(data[offset : offset+subPathLength])
		offset += subPathLength

		tag = &SoftObjectProperty{
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
	default:
		if offset < uint32(len(data)-1) {
			/*
				fmt.Println()
				fmt.Println(tagData)
				fmt.Println(propertyType, ": MAY BE UNREAD DATA ("+strconv.Itoa(len(data[offset:]))+"):")
				fmt.Println(utils.HexDump(data[offset:]))
			*/

			// TODO Read unknown cases
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

	branchLength := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	branch := string(data[offset : offset+branchLength])
	offset += branchLength

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
