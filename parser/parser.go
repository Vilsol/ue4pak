package parser

import (
	"compress/zlib"
	"fmt"
	"github.com/spf13/viper"
)

type PakParser struct {
	reader     PakReader
	tracker    *readTracker
	preload    []byte
	baseReader PakReader
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

func (parser *PakParser) TrackRead() *readTracker {
	parser.tracker = &readTracker{
		child: parser.tracker,
	}

	return parser.tracker
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

func (parser *PakParser) StartCompression(method uint32) {
	if method != 1 {
		panic(fmt.Sprintf("unknown compression method: %d", method))
	}

	parser.baseReader = parser.reader

	zlibReader, err := zlib.NewReader(parser.baseReader)
	if err != nil {
		panic(err)
	}

	parser.reader = &PakZlibReader{
		Reader: zlibReader,
	}
}

func (parser *PakParser) StopCompression() {
	parser.reader = parser.baseReader
}
