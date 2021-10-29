package parser

import (
	"encoding/binary"
)

const INDEX_NONE = int64(-1)

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
			panic("Could not find magic bytes in pak!")
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
	parser.Preload(int32(pakFooter.IndexSize))

	mountPoint := parser.ReadString()
	recordCount := parser.ReadInt32()

	pakIndex := &FPakIndex{
		MountPoint: mountPoint,
		Records:    make([]*FPakEntry, recordCount),
	}

	if pakFooter.Version >= 10 {
		parser.DecodePakEntries(pakIndex, pakFooter)
	} else {
		parser.DecodeLegacyPakEntries(pakIndex, pakFooter)
	}

	return &PakFile{
		Footer: pakFooter,
		Index:  pakIndex,
	}
}

func (parser *PakParser) DecodePakEntries(pakIndex *FPakIndex, pakFooter *FPakInfo) {
	parser.ReadUint64() // PathHashSeed

	bReaderHasPathHashIndex := parser.ReadInt32() == 1
	//PathHashIndexOffset := INDEX_NONE
	//PathHashIndexSize := int64(0)
	//var PathHashIndexHash []byte
	if bReaderHasPathHashIndex {
		parser.ReadInt64() // PathHashIndexOffset
		parser.ReadInt64() // PathHashIndexSize
		parser.Read(20)    // PathHashIndexHash
	}

	bReaderHasFullDirectoryIndex := parser.ReadInt32() == 1
	FullDirectoryIndexOffset := INDEX_NONE
	FullDirectoryIndexSize := int64(0)
	//var FullDirectoryIndexHash []byte
	if bReaderHasFullDirectoryIndex {
		FullDirectoryIndexOffset = parser.ReadInt64()
		FullDirectoryIndexSize = parser.ReadInt64()
		parser.Read(20) // FullDirectoryIndexHash
	}

	parser.ReadInt32() // EncodedPakEntryLength

	encodedIndex := make(map[int32]*FPakEntry)

	tracker := parser.TrackRead()
	for i := 0; i < len(pakIndex.Records); i++ {
		position := tracker.bytesRead
		entry := &FPakEntry{}

		// Grab the big bitfield value:
		// Bit 31 = Offset 32-bit safe?
		// Bit 30 = Uncompressed size 32-bit safe?
		// Bit 29 = Size 32-bit safe?
		// Bits 28-23 = Compression method
		// Bit 22 = Encrypted
		// Bits 21-6 = Compression blocks count
		// Bits 5-0 = Compression block size
		value := parser.ReadUint32()

		entry.CompressionMethod = value >> 23 & 0x3f

		bIsOffset32BitSafe := (value & (1 << 31)) != 0
		if bIsOffset32BitSafe {
			entry.FileOffset = int64(parser.ReadUint32())
		} else {
			entry.FileOffset = parser.ReadInt64()
		}

		bIsUncompressedSize32BitSafe := (value & (1 << 30)) != 0
		if bIsUncompressedSize32BitSafe {
			entry.UncompressedSize = int64(parser.ReadUint32())
		} else {
			entry.UncompressedSize = parser.ReadInt64()
		}

		if entry.CompressionMethod != 0 {
			bIsSize32BitSafe := (value & (1 << 29)) != 0
			if bIsSize32BitSafe {
				entry.FileSize = int64(parser.ReadUint32())
			} else {
				entry.FileSize = parser.ReadInt64()
			}
		} else {
			entry.FileSize = entry.UncompressedSize
		}

		entry.IsEncrypted = (value & (1 << 22)) != 0

		CompressionBlocksCount := (value >> 6) & 0xffff
		entry.CompressionBlocks = make([]*FPakCompressedBlock, CompressionBlocksCount)

		entry.CompressionBlockSize = 0
		if CompressionBlocksCount > 0 {
			if entry.UncompressedSize < 65536 {
				entry.CompressionBlockSize = uint32(entry.UncompressedSize)
			} else {
				entry.CompressionBlockSize = (value & 0x3f) << 11
			}
		}

		// TODO
		//baseOffset := entry.FileOffset
		//if pakFooter.Version >= 5 {
		//	baseOffset = 0
		//}

		if len(entry.CompressionBlocks) == 1 && !entry.IsEncrypted {
			panic("TODO") // TODO
		} else if len(entry.CompressionBlocks) > 0 {
			panic("TODO") // TODO
		}

		encodedIndex[position] = entry
		pakIndex.Records[i] = entry
	}
	parser.UnTrackRead()

	FilesNum := parser.ReadInt32()

	if FilesNum > 0 {
		pakIndex.Records = make([]*FPakEntry, FilesNum)

		for i := int32(0); i < FilesNum; i++ {
			parser.DecodeFPakEntry(pakIndex.Records[i], pakFooter.Version)
		}
	}

	//bWillUseFullDirectoryIndex := false
	//bWillUsePathHashIndex := false
	bReadFullDirectoryIndex := false
	if bReaderHasPathHashIndex && bReaderHasFullDirectoryIndex {
		//bWillUseFullDirectoryIndex = IsPakKeepFullDirectory()
		//bWillUsePathHashIndex = !bWillUseFullDirectoryIndex
		bWantToReadFullDirectoryIndex := IsPakKeepFullDirectory()
		bReadFullDirectoryIndex = bReaderHasFullDirectoryIndex && bWantToReadFullDirectoryIndex
	} else if bReaderHasPathHashIndex {
		//bWillUsePathHashIndex = true
		//bWillUseFullDirectoryIndex = false
		bReadFullDirectoryIndex = false
	} else if bReaderHasFullDirectoryIndex {
		//bWillUsePathHashIndex = false
		//bWillUseFullDirectoryIndex = true
		bReadFullDirectoryIndex = true
	}

	//bHasFullDirectoryIndex := false

	FDirectoryIndex := make(map[string]map[string]FPakEntryLocation)

	if !bReadFullDirectoryIndex {
		// TODO
		// PathHashIndexReader << DirectoryIndex
		//bHasFullDirectoryIndex = false
		panic("TODO") // TODO
	} else {
		parser.Seek(FullDirectoryIndexOffset, 0)
		parser.Preload(int32(FullDirectoryIndexSize))

		directoryCount := parser.ReadInt32()
		for i := int32(0); i < directoryCount; i++ {
			directoryName := parser.ReadString()
			fileCount := parser.ReadInt32()
			FPakDirectory := make(map[string]FPakEntryLocation, fileCount)
			for j := int32(0); j < fileCount; j++ {
				fileName := parser.ReadString()
				location := FPakEntryLocation{
					Index: parser.ReadInt32(),
				}

				if entry, ok := encodedIndex[location.Index]; ok {
					// Strip null byte from end of directory name
					entry.FileName = directoryName[:len(directoryName)-1] + fileName
				}

				FPakDirectory[fileName] = location
			}
			FDirectoryIndex[directoryName] = FPakDirectory
		}

		//bHasFullDirectoryIndex = true
	}
}

func IsPakKeepFullDirectory() bool {
	return true
}

func (parser *PakParser) DecodeLegacyPakEntries(pakIndex *FPakIndex, pakFooter *FPakInfo) {
	for i := 0; i < len(pakIndex.Records); i++ {
		entry := &FPakEntry{}
		parser.DecodeFPakEntry(entry, pakFooter.Version)
		pakIndex.Records[i] = entry
	}
}

func (parser *PakParser) DecodeFPakEntry(entry *FPakEntry, version uint32) {
	entry.FileName = parser.ReadString()
	entry.FileOffset = parser.ReadInt64()
	entry.FileSize = parser.ReadInt64()
	entry.UncompressedSize = parser.ReadInt64()

	if version <= 8 {
		entry.CompressionMethod = uint32(parser.Read(1)[0])
	} else {
		entry.CompressionMethod = uint32(parser.ReadInt32())
	}

	if version <= 1 {
		entry.Timestamp = parser.ReadUint64()
	}

	entry.DataSHA1Hash = parser.Read(20)

	if version >= 3 {
		if entry.CompressionMethod != 0 {
			blockCount := parser.ReadUint32()

			entry.CompressionBlocks = make([]*FPakCompressedBlock, blockCount)

			for j := 0; j < len(entry.CompressionBlocks); j++ {
				entry.CompressionBlocks[j] = &FPakCompressedBlock{
					StartOffset: parser.ReadUint64(),
					EndOffset:   parser.ReadUint64(),
				}
			}
		}

		entry.IsEncrypted = parser.Read(1)[0] > 0
		entry.CompressionBlockSize = parser.ReadUint32()
	}

	if version == 4 {
		// TODO ???
	}

	if version >= 9 {
		// TODO Unknown bytes
		// parser.Read(3)
	}

}
