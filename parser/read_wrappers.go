package parser

type PakReader interface {
	Seek(offset int64, whence int) (ret int64, err error)
	Read(b []byte) (n int, err error)
}

type PakByteReader struct {
	PakReader

	Bytes  []byte
	Offset int64
}

func (reader *PakByteReader) Seek(offset int64, whence int) (ret int64, err error) {
	if whence == 0 {
		reader.Offset = offset
	} else if whence == 1 {
		reader.Offset += offset
	} else if whence == 2 {
		reader.Offset = int64(len(reader.Bytes)) + offset
	}

	return reader.Offset, nil
}

func (reader *PakByteReader) Read(b []byte) (n int, err error) {
	copied := copy(b, reader.Bytes[reader.Offset:])
	return copied, nil
}