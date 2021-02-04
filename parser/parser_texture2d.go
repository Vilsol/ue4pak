package parser

import (
	log "github.com/sirupsen/logrus"
	"github.com/spate/glimage"
	"github.com/x448/float16"
	"image"
	"strings"
)

type Texture2D struct {
	Cooked   uint32
	Textures []*FTexturePlatformData
}

type FTexturePlatformData struct {
	SizeX       int32               `json:"size_x"`
	SizeY       int32               `json:"size_y"`
	NumSlices   int32               `json:"num_slices"`
	PixelFormat string              `json:"pixel_format"`
	FirstMip    int32               `json:"first_mip"`
	Mips        []*FTexture2DMipMap `json:"-"`
	IsVirtual   bool                `json:"is_virtual"`
}

type FTexture2DMipMap struct {
	Data  *FByteBulkData
	SizeX int32
	SizeY int32
	SizeZ int32
}

type FByteBulkData struct {
	Header *FByteBulkDataHeader
	Data   []byte
}

type FByteBulkDataHeader struct {
	BulkDataFlags int32
	ElementCount  int32
	SizeOnDisk    int32
	OffsetInFile  int64
}

func (parser *PakParser) ReadFTexturePlatformData(bulkOffset int64) *FTexturePlatformData {
	data := &FTexturePlatformData{
		SizeX:       parser.ReadInt32(),
		SizeY:       parser.ReadInt32(),
		NumSlices:   parser.ReadInt32(),
		PixelFormat: parser.ReadString(),
		FirstMip:    parser.ReadInt32(),
	}

	length := parser.ReadUint32()
	data.Mips = make([]*FTexture2DMipMap, length)

	for i := uint32(0); i < length; i++ {
		data.Mips[i] = parser.ReadFTexture2DMipMap(bulkOffset)
	}

	return data
}

func (parser *PakParser) ReadFTexture2DMipMap(bulkOffset int64) *FTexture2DMipMap {
	cooked := parser.ReadInt32()

	mipMap := &FTexture2DMipMap{
		Data:  parser.ReadFByteBulkData(bulkOffset),
		SizeX: parser.ReadInt32(),
		SizeY: parser.ReadInt32(),
		SizeZ: parser.ReadInt32(),
	}

	if cooked != 1 {
		log.Errorf("Uncooked FTexture2DMipMap: %s", parser.ReadString())
		return nil
	}

	return mipMap
}

func (parser *PakParser) ReadFByteBulkData(bulkOffset int64) *FByteBulkData {
	header := parser.ReadFByteBulkDataHeader()

	var data []byte

	if header.BulkDataFlags&0x0040 != 0 {
		data = parser.Read(header.ElementCount)
	}

	if header.BulkDataFlags&0x0100 != 0 {
		panic("TODO") // TODO
	}

	return &FByteBulkData{
		Header: header,
		Data:   data,
	}
}

func (parser *PakParser) ReadFByteBulkDataHeader() *FByteBulkDataHeader {
	return &FByteBulkDataHeader{
		BulkDataFlags: parser.ReadInt32(),
		ElementCount:  parser.ReadInt32(),
		SizeOnDisk:    parser.ReadInt32(),
		OffsetInFile:  parser.ReadInt64(),
	}
}

func (texture *Texture2D) ToImage() image.Image {
	mipMap := texture.Textures[0].Mips[0]

	switch strings.Trim(texture.Textures[0].PixelFormat, "\x00") {
	case "PF_DXT1":
		return DecodeDXT1(mipMap.Data.Data, mipMap.SizeX, mipMap.SizeY)
	case "PF_DXT3":
		return DecodeDXT3(mipMap.Data.Data, mipMap.SizeX, mipMap.SizeY)
	case "PF_DXT5":
		return DecodeDXT5(mipMap.Data.Data, mipMap.SizeX, mipMap.SizeY)
	case "PF_B8G8R8A8":
		return DecodeBGRA(mipMap.Data.Data, mipMap.SizeX, mipMap.SizeY)
	case "PF_G8":
		return DecodeG8(mipMap.Data.Data, mipMap.SizeX, mipMap.SizeY)
	case "PF_FloatRGBA":
		return DecodeFloatRGBA(mipMap.Data.Data, mipMap.SizeX, mipMap.SizeY)
	default:
		log.Errorf("Unknown Texture2D pixel format: %s", strings.Trim(texture.Textures[0].PixelFormat, "\x00"))
		return nil
	}
}

func DecodeDXT1(data []byte, width int32, height int32) image.Image {
	img := glimage.NewDxt1(image.Rect(0, 0, int(width), int(height)))
	img.Pix = data[:]
	return img
}

func DecodeDXT3(data []byte, width int32, height int32) image.Image {
	img := glimage.NewDxt3(image.Rect(0, 0, int(width), int(height)))
	img.Pix = data[:]
	return img
}

func DecodeDXT5(data []byte, width int32, height int32) image.Image {
	img := glimage.NewDxt5(image.Rect(0, 0, int(width), int(height)))
	img.Pix = data[:]
	return img
}

func DecodeBGRA(data []byte, width int32, height int32) image.Image {
	img := glimage.NewBGRA(image.Rect(0, 0, int(width), int(height)))
	img.Pix = data[:]
	return img
}

func DecodeG8(data []byte, width int32, height int32) image.Image {
	img := image.NewGray(image.Rect(0, 0, int(width), int(height)))
	img.Pix = data[:]
	return img
}

func DecodeFloatRGBA(data []byte, width int32, height int32) image.Image {
	newData := make([]byte, width*height*4)

	for i := int32(0); i < width*height*4; i++ {
		newData[i] = uint8(float16.Frombits(uint16(data[i*2+1])<<8|uint16(data[i*2])).Float32() * 255)
	}

	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	img.Pix = newData

	return img
}
