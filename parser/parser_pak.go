package parser

import (
	"encoding/binary"
)

func (parser *PakParser) Parse() (*PakFile, error) {
	// Find magic number
	magicOffset := int64(-44)

	for {
		_, err := parser.Seek(magicOffset, 2)
		if err != nil {
			return nil, err
		}

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
	_, err := parser.Seek(magicOffset, 2)
	if err != nil {
		return nil, err
	}

	footer := parser.Read(int32(magicOffset * -1))

	pakFooter := &FPakInfo{
		Magic:         binary.LittleEndian.Uint32(footer[0:4]),
		Version:       binary.LittleEndian.Uint32(footer[4:8]),
		IndexOffset:   binary.LittleEndian.Uint64(footer[8:16]),
		IndexSize:     binary.LittleEndian.Uint64(footer[16:24]),
		IndexSHA1Hash: footer[24:44],
	}

	// Seek and read the index of the file
	_, err = parser.Seek(int64(pakFooter.IndexOffset), 0)
	if err != nil {
		return nil, err
	}

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

		if pakFooter.Version >= 9 {
			// TODO Unknown bytes
			parser.Read(3)
		}
	}

	return &PakFile{
		Footer: pakFooter,
		Index:  pakIndex,
	}, nil
}
