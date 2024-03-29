package parser

import (
	"encoding/json"
	"github.com/spf13/viper"
)

var mapPropertyTypeOverrides = map[string]*MapProperty{
	"BindingIdToReferences": {
		KeyType:   "Guid",
		ValueType: "LevelSequenceBindingReferenceArray",
	},
	"Tracks": {
		KeyType:   "MovieSceneTrackIdentifier",
		ValueType: "MovieSceneEvaluationTrack",
	},
	"SubTemplateSerialNumbers": {
		KeyType:   "MovieSceneSequenceID",
		ValueType: "UInt32Property",
	},
	"SubSequences": {
		KeyType:   "MovieSceneSequenceID",
		ValueType: "MovieSceneSubSequenceData",
	},
	"Hierarchy": {
		KeyType:   "MovieSceneSequenceID",
		ValueType: "MovieSceneSequenceHierarchyNode",
	},
	"TrackSignatureToTrackIdentifier": {
		KeyType:   "Guid",
		ValueType: "MovieSceneTrackIdentifier",
	},
	"SubSectionRanges": {
		KeyType:   "Guid",
		ValueType: "MovieSceneFrameRange",
	},
}

func RegisterMapPropertyOverride(name string, override *MapProperty) {
	mapPropertyTypeOverrides[name] = override
}

type PakEntrySet struct {
	ExportRecord *FPakEntry           `json:"export_record"`
	Summary      *FPackageFileSummary `json:"summary"`
	Exports      []PakExportSet       `json:"exports"`
}

type PakExportSet struct {
	Export *FObjectExport `json:"export"`
	Data   *ExportData    `json:"data"`
}

type FPakInfo struct {
	Magic           uint32 `json:"magic"`
	Version         uint32 `json:"version"`
	IndexOffset     uint64 `json:"index_offset"`
	IndexSize       uint64 `json:"index_size"`
	IndexSHA1Hash   []byte `json:"index_sha_1_hash"`
	CompressionType string `json:"compression_type"`
}

type FPakIndex struct {
	MountPoint string       `json:"mount_point"`
	Records    []*FPakEntry `json:"records"`
}

type FPakEntry struct {
	FileName          string `json:"file_name"`
	FileOffset        int64  `json:"file_offset"`
	FileSize          int64  `json:"file_size"`
	UncompressedSize  int64  `json:"uncompressed_size"`
	CompressionMethod uint32 `json:"compression_method"`

	// Only version <= 1
	Timestamp uint64 `json:"timestamp"`

	DataSHA1Hash []byte `json:"data_sha_1_hash"`

	// Only version >= 3
	// Only compressed
	CompressionBlocks []*FPakCompressedBlock `json:"compression_blocks"`

	IsEncrypted          bool   `json:"is_encrypted"`
	CompressionBlockSize uint32 `json:"compression_block_size"`
}

type FPakCompressedBlock struct {
	StartOffset uint64 `json:"start_offset"`
	EndOffset   uint64 `json:"end_offset"`
}

type PakFile struct {
	Footer *FPakInfo  `json:"footer"`
	Index  *FPakIndex `json:"index"`
}

type FNameEntrySerialized struct {
	Name                  string `json:"name"`
	NonCasePreservingHash uint16 `json:"non_case_preserving_hash"`
	CasePreservingHash    uint16 `json:"case_preserving_hash"`
}

type FObjectImport struct {
	ClassPackage string         `json:"class_package"`
	ClassName    string         `json:"class_name"`
	OuterIndex   int32          `json:"outer_index"`
	ObjectName   string         `json:"object_name"`
	OuterPackage *FPackageIndex `json:"outer_package"`
}

func (m *FObjectImport) MarshalJSON() ([]byte, error) {
	ex := &struct {
		ClassPackage string         `json:"class_package"`
		ClassName    string         `json:"class_name"`
		OuterIndex   int32          `json:"outer_index"`
		ObjectName   string         `json:"object_name"`
		OuterPackage *FPackageIndex `json:"outer_package"`
	}{
		ClassPackage: m.ClassPackage,
		ClassName:    m.ClassName,
		OuterIndex:   m.OuterIndex,
		ObjectName:   m.ObjectName,
	}

	if viper.GetBool("with-index") {
		ex.OuterPackage = m.OuterPackage
	}

	return json.Marshal(&ex)
}

type FPackageIndex struct {
	Index     int32       `json:"index"`
	Reference interface{} `json:"reference"`
}

type FObjectExport struct {
	ClassIndex                                   *FPackageIndex `json:"class_index"`
	SuperIndex                                   *FPackageIndex `json:"super_index"`
	TemplateIndex                                *FPackageIndex `json:"template_index"`
	OuterIndex                                   *FPackageIndex `json:"outer_index"`
	ObjectName                                   string         `json:"object_name"`
	Save                                         uint32         `json:"save"`
	SerialSize                                   int64          `json:"serial_size"`
	SerialOffset                                 int64          `json:"serial_offset"`
	ForcedExport                                 bool           `json:"forced_export"`
	NotForClient                                 bool           `json:"not_for_client"`
	NotForServer                                 bool           `json:"not_for_server"`
	PackageGuid                                  *FGuid         `json:"package_guid"`
	PackageFlags                                 uint32         `json:"package_flags"`
	NotAlwaysLoadedForEditorGame                 bool           `json:"not_always_loaded_for_editor_game"`
	IsAsset                                      bool           `json:"is_asset"`
	FirstExportDependency                        int32          `json:"first_export_dependency"`
	SerializationBeforeSerializationDependencies bool           `json:"serialization_before_serialization_dependencies"`
	CreateBeforeSerializationDependencies        bool           `json:"create_before_serialization_dependencies"`
	SerializationBeforeCreateDependencies        bool           `json:"serialization_before_create_dependencies"`
	CreateBeforeCreateDependencies               bool           `json:"create_before_create_dependencies"`
}

func (m *FObjectExport) MarshalJSON() ([]byte, error) {
	ex := &struct {
		ClassIndex                                   *FPackageIndex `json:"class_index"`
		SuperIndex                                   *FPackageIndex `json:"super_index"`
		TemplateIndex                                *FPackageIndex `json:"template_index"`
		OuterIndex                                   *FPackageIndex `json:"outer_index"`
		ObjectName                                   string         `json:"object_name"`
		Save                                         uint32         `json:"save"`
		SerialSize                                   int64          `json:"serial_size"`
		SerialOffset                                 int64          `json:"serial_offset"`
		ForcedExport                                 bool           `json:"forced_export"`
		NotForClient                                 bool           `json:"not_for_client"`
		NotForServer                                 bool           `json:"not_for_server"`
		PackageGuid                                  *FGuid         `json:"package_guid"`
		PackageFlags                                 uint32         `json:"package_flags"`
		NotAlwaysLoadedForEditorGame                 bool           `json:"not_always_loaded_for_editor_game"`
		IsAsset                                      bool           `json:"is_asset"`
		FirstExportDependency                        int32          `json:"first_export_dependency"`
		SerializationBeforeSerializationDependencies bool           `json:"serialization_before_serialization_dependencies"`
		CreateBeforeSerializationDependencies        bool           `json:"create_before_serialization_dependencies"`
		SerializationBeforeCreateDependencies        bool           `json:"serialization_before_create_dependencies"`
		CreateBeforeCreateDependencies               bool           `json:"create_before_create_dependencies"`
	}{
		ObjectName:                   m.ObjectName,
		Save:                         m.Save,
		SerialSize:                   m.SerialSize,
		SerialOffset:                 m.SerialOffset,
		ForcedExport:                 m.ForcedExport,
		NotForClient:                 m.NotForClient,
		NotForServer:                 m.NotForServer,
		PackageGuid:                  m.PackageGuid,
		PackageFlags:                 m.PackageFlags,
		NotAlwaysLoadedForEditorGame: m.NotAlwaysLoadedForEditorGame,
		IsAsset:                      m.IsAsset,
		FirstExportDependency:        m.FirstExportDependency,
		SerializationBeforeSerializationDependencies: m.SerializationBeforeSerializationDependencies,
		CreateBeforeSerializationDependencies:        m.CreateBeforeSerializationDependencies,
		SerializationBeforeCreateDependencies:        m.SerializationBeforeCreateDependencies,
		CreateBeforeCreateDependencies:               m.CreateBeforeCreateDependencies,
	}

	if viper.GetBool("with-index") {
		ex.ClassIndex = m.ClassIndex
		ex.SuperIndex = m.SuperIndex
		ex.TemplateIndex = m.TemplateIndex
		ex.OuterIndex = m.OuterIndex
	}

	return json.Marshal(&ex)
}

type FPackageFileSummary struct {
	Record *FPakEntry `json:"record"`

	Tag                         int32                   `json:"tag"`
	LegacyFileVersion           int32                   `json:"legacy_file_version"`
	LegacyUE3Version            int32                   `json:"legacy_ue_3_version"`
	FileVersionUE4              int32                   `json:"file_version_ue_4"`
	FileVersionLicenseeUE4      int32                   `json:"file_version_licensee_ue_4"`
	TotalHeaderSize             int32                   `json:"total_header_size"`
	FolderName                  string                  `json:"folder_name"`
	PackageFlags                uint32                  `json:"package_flags"`
	NameOffset                  int32                   `json:"name_offset"`
	GatherableTextDataCount     int32                   `json:"gatherable_text_data_count"`
	GatherableTextDataOffset    int32                   `json:"gatherable_text_data_offset"`
	ExportOffset                int32                   `json:"export_offset"`
	ImportOffset                int32                   `json:"import_offset"`
	DependsOffset               int32                   `json:"depends_offset"`
	StringAssetReferencesCount  int32                   `json:"string_asset_references_count"`
	StringAssetReferencesOffset int32                   `json:"string_asset_references_offset"`
	SearchableNamesOffset       int32                   `json:"searchable_names_offset"`
	ThumbnailTableOffset        int32                   `json:"thumbnail_table_offset"`
	GUID                        *FGuid                  `json:"guid"`
	Generations                 []*FGenerationInfo      `json:"generations"`
	SavedByEngineVersion        *FEngineVersion         `json:"saved_by_engine_version"`
	CompatibleWithEngineVersion *FEngineVersion         `json:"compatible_with_engine_version"`
	CompressionFlags            uint32                  `json:"compression_flags"`
	CompressedChunks            []*FCompressedChunk     `json:"compressed_chunks"`
	PackageSource               uint32                  `json:"package_source"`
	AdditionalPackagesToCook    []string                `json:"additional_packages_to_cook"`
	AssetRegistryDataOffset     int32                   `json:"asset_registry_data_offset"`
	BulkDataStartOffset         int32                   `json:"bulk_data_start_offset"`
	WorldTileInfoDataOffset     int32                   `json:"world_tile_info_data_offset"`
	ChunkIds                    []int32                 `json:"chunk_ids"`
	PreloadDependencyCount      int32                   `json:"preload_dependency_count"`
	PreloadDependencyOffset     int32                   `json:"preload_dependency_offset"`
	Names                       []*FNameEntrySerialized `json:"names"`
	Imports                     []*FObjectImport        `json:"imports"`
	Exports                     []*FObjectExport        `json:"exports"`
}

func (m *FPackageFileSummary) MarshalJSON() ([]byte, error) {
	ex := &struct {
		Record                      *FPakEntry              `json:"record"`
		Tag                         int32                   `json:"tag"`
		LegacyFileVersion           int32                   `json:"legacy_file_version"`
		LegacyUE3Version            int32                   `json:"legacy_ue_3_version"`
		FileVersionUE4              int32                   `json:"file_version_ue_4"`
		FileVersionLicenseeUE4      int32                   `json:"file_version_licensee_ue_4"`
		TotalHeaderSize             int32                   `json:"total_header_size"`
		FolderName                  string                  `json:"folder_name"`
		PackageFlags                uint32                  `json:"package_flags"`
		NameOffset                  int32                   `json:"name_offset"`
		GatherableTextDataCount     int32                   `json:"gatherable_text_data_count"`
		GatherableTextDataOffset    int32                   `json:"gatherable_text_data_offset"`
		ExportOffset                int32                   `json:"export_offset"`
		ImportOffset                int32                   `json:"import_offset"`
		DependsOffset               int32                   `json:"depends_offset"`
		StringAssetReferencesCount  int32                   `json:"string_asset_references_count"`
		StringAssetReferencesOffset int32                   `json:"string_asset_references_offset"`
		SearchableNamesOffset       int32                   `json:"searchable_names_offset"`
		ThumbnailTableOffset        int32                   `json:"thumbnail_table_offset"`
		GUID                        *FGuid                  `json:"guid"`
		Generations                 []*FGenerationInfo      `json:"generations"`
		SavedByEngineVersion        *FEngineVersion         `json:"saved_by_engine_version"`
		CompatibleWithEngineVersion *FEngineVersion         `json:"compatible_with_engine_version"`
		CompressionFlags            uint32                  `json:"compression_flags"`
		CompressedChunks            []*FCompressedChunk     `json:"compressed_chunks"`
		PackageSource               uint32                  `json:"package_source"`
		AdditionalPackagesToCook    []string                `json:"additional_packages_to_cook"`
		AssetRegistryDataOffset     int32                   `json:"asset_registry_data_offset"`
		BulkDataStartOffset         int32                   `json:"bulk_data_start_offset"`
		WorldTileInfoDataOffset     int32                   `json:"world_tile_info_data_offset"`
		ChunkIds                    []int32                 `json:"chunk_ids"`
		PreloadDependencyCount      int32                   `json:"preload_dependency_count"`
		PreloadDependencyOffset     int32                   `json:"preload_dependency_offset"`
		Names                       []*FNameEntrySerialized `json:"names"`
		Imports                     []*FObjectImport        `json:"imports"`
		Exports                     []*FObjectExport        `json:"exports"`
	}{
		Record:                      m.Record,
		Tag:                         m.Tag,
		LegacyFileVersion:           m.LegacyFileVersion,
		LegacyUE3Version:            m.LegacyUE3Version,
		FileVersionUE4:              m.FileVersionUE4,
		FileVersionLicenseeUE4:      m.FileVersionLicenseeUE4,
		TotalHeaderSize:             m.TotalHeaderSize,
		FolderName:                  m.FolderName,
		PackageFlags:                m.PackageFlags,
		NameOffset:                  m.NameOffset,
		GatherableTextDataCount:     m.GatherableTextDataCount,
		GatherableTextDataOffset:    m.GatherableTextDataOffset,
		ExportOffset:                m.ExportOffset,
		ImportOffset:                m.ImportOffset,
		DependsOffset:               m.DependsOffset,
		StringAssetReferencesCount:  m.StringAssetReferencesCount,
		StringAssetReferencesOffset: m.StringAssetReferencesOffset,
		SearchableNamesOffset:       m.SearchableNamesOffset,
		ThumbnailTableOffset:        m.ThumbnailTableOffset,
		GUID:                        m.GUID,
		Generations:                 m.Generations,
		SavedByEngineVersion:        m.SavedByEngineVersion,
		CompatibleWithEngineVersion: m.CompatibleWithEngineVersion,
		CompressionFlags:            m.CompressionFlags,
		CompressedChunks:            m.CompressedChunks,
		PackageSource:               m.PackageSource,
		AdditionalPackagesToCook:    m.AdditionalPackagesToCook,
		AssetRegistryDataOffset:     m.AssetRegistryDataOffset,
		BulkDataStartOffset:         m.BulkDataStartOffset,
		WorldTileInfoDataOffset:     m.WorldTileInfoDataOffset,
		ChunkIds:                    m.ChunkIds,
		PreloadDependencyCount:      m.PreloadDependencyCount,
		PreloadDependencyOffset:     m.PreloadDependencyOffset,
		Imports:                     m.Imports,
		Exports:                     m.Exports,
	}

	if viper.GetBool("with-names") {
		ex.Names = m.Names
	}

	return json.Marshal(&ex)
}

type UObject struct {
	ExportType string          `json:"export_type"`
	Properties []*FPropertyTag `json:"properties"`
}

type FPropertyTag struct {
	Name         string      `json:"name"`
	PropertyType string      `json:"property_type"`
	TagData      interface{} `json:"tag_data"`
	Size         int32       `json:"size"`
	ArrayIndex   int32       `json:"array_index"`
	PropertyGuid *FGuid      `json:"property_guid"`
	Tag          interface{} `json:"tag"`
}

type StructProperty struct {
	Type string `json:"type"`
	Guid *FGuid `json:"guid"`
}

type FSoftObjectPath struct {
	AssetPathName string `json:"asset_path_name"`
	SubPath       string `json:"sub_path"`
}

type FEngineVersion struct {
	Major      uint16 `json:"major"`
	Minor      uint16 `json:"minor"`
	Patch      uint16 `json:"patch"`
	ChangeList uint32 `json:"change_list"`
	Branch     string `json:"branch"`
}

type FGenerationInfo struct {
	ExportCount int32 `json:"export_count"`
	NameCount   int32 `json:"name_count"`
}

type FCompressedChunk struct {
	UncompressedOffset int32 `json:"uncompressed_offset"`
	UncompressedSize   int32 `json:"uncompressed_size"`
	CompressedOffset   int32 `json:"compressed_offset"`
	CompressedSize     int32 `json:"compressed_size"`
}

type MapProperty struct {
	KeyType   string `json:"key_type"`
	ValueType string `json:"value_type"`
}

type UInterfaceProperty struct {
	InterfaceNumber uint32 `json:"interface_number"`
}

type FText struct {
	Flags        uint32 `json:"flags"`
	HistoryType  int8   `json:"history_type"`
	Namespace    string `json:"namespace"`
	Key          string `json:"key"`
	SourceString string `json:"source_string"`
}

type FScriptDelegate struct {
	Object int32  `json:"object"`
	Name   string `json:"name"`
}

type ArrayStructProperty struct {
	InnerTagData *FPropertyTag `json:"inner_tag_data"`
	Properties   interface{}   `json:"properties"`
}

type MapPropertyEntry struct {
	Key   interface{} `json:"key"`
	Value interface{} `json:"value"`
}

type ExportData struct {
	Properties []*FPropertyTag `json:"properties"`
	Data       interface{}     `json:"data"`
}

type FPakEntryLocation struct {
	Index int32 `json:"index"`
}

func (pakInfo *FPakInfo) HeaderSize() uint64 {
	if pakInfo.Version < 8 {
		return 53
	}

	return 50
}

func (index *FPackageIndex) ObjectName() *string {
	classReference := index.Reference

	if classReference == nil {
		return nil
	}

	if ref, ok := classReference.(*FObjectImport); ok {
		return &ref.ObjectName
	} else if ref, ok := classReference.(*FObjectExport); ok {
		return &ref.ObjectName
	}

	return nil
}

func (index *FPackageIndex) ClassName() *string {
	classReference := index.Reference

	if classReference == nil {
		return nil
	}

	if ref, ok := classReference.(*FObjectImport); ok {
		return &ref.ClassName
	} else if ref, ok := classReference.(*FObjectExport); ok {
		return ref.ClassIndex.ObjectName()
	}

	return nil
}
