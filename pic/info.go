package pic

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"net/http"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"

	"github.com/pkg/errors"
)

type PicInfo struct {
	Format string // 图像数据格式，取值应符合GA/T 1286-2015中5.3.5要求
	Width  int    // 图像水平像素数
	Height int    // 图像垂直像素数
	Depth  int    // 图像数据位深度，取值应符合GA/T 1286-2015中5.3.5要求
	Length int    // 图像数据长度，取值应符合GA/T 1286-2015中5.3.5要求
}

func GetPicInfo(b []byte) (picInfo PicInfo, err error) {
	// base.IKIRD_GetIrisFeature(b)

	// 获取图片格式
	f := http.DetectContentType(b)
	if strings.HasPrefix(f, "image/") {
		picInfo.Format = f[6:]
	}

	// 图片长度
	picInfo.Length = len(b)

	// 图片类型
	switch picInfo.Format {
	case "jpeg":
		picInfo.Width, picInfo.Height, err = getJpegHw(b)
	case "png":
		picInfo.Width, picInfo.Height, err = GetPngWidthHeight(b)
	case "bmp":
		picInfo.Width, picInfo.Height, err = GetBmpWidthHeight(b)
	default:
		err = errors.New("getPicInfo: unknown format")
	}
	if err != nil {
		return picInfo, errors.WithStack(err)
	}

	// 图片深度
	img, _, err := image.DecodeConfig(bytes.NewBuffer(b))
	if err != nil {
		return picInfo, errors.WithStack(err)
	}
	// Get the color model and bit depth of the image
	switch m := img.ColorModel; m {
	case color.RGBAModel:
		picInfo.Depth = 32
	case color.NRGBAModel:
		picInfo.Depth = 32
	case color.RGBA64Model:
		picInfo.Depth = 64
	case color.NRGBA64Model:
		picInfo.Depth = 64
	case color.GrayModel:
		picInfo.Depth = 8
	case color.Gray16Model:
		picInfo.Depth = 16
	case color.CMYKModel:
		picInfo.Depth = 32
	case color.YCbCrModel:
		picInfo.Depth = 24
	case color.NYCbCrAModel:
		picInfo.Depth = 32
	case color.AlphaModel:
		picInfo.Depth = 8
	case color.Alpha16Model:
		picInfo.Depth = 16
	default:
		picInfo.Depth = 8
	}

	return

}

func getJpegHw(b []byte) (w, h int, err error) {
	var offset int
	imgByteLen := len(b)
	for i := 0; i < imgByteLen-1; i++ {
		if b[i] != 0xff {
			continue
		}
		if b[i+1] == 0xC0 || b[i+1] == 0xC1 || b[i+1] == 0xC2 {
			offset = i
			break
		}
	}
	offset += 5
	if offset >= imgByteLen {
		return 0, 0, errors.New("getJpegHw: offset out of range")
	}
	height := int(b[offset])<<8 + int(b[offset+1])
	width := int(b[offset+2])<<8 + int(b[offset+3])
	return width, height, nil
}

// 获取 PNG 图片的宽高
func GetPngWidthHeight(b []byte) (w, h int, err error) {
	pngHeader := "\x89PNG\r\n\x1a\n"
	if string(b[:len(pngHeader)]) != pngHeader {
		return 0, 0, errors.New("GetPngWidthHeight: not a png file")
	}
	offset := 12
	if string(b[offset:offset+4]) != "IHDR" {
		return 0, 0, errors.New("GetPngWidthHeight: IHDR not found")
	}
	offset += 4
	width := int(binary.BigEndian.Uint32(b[offset : offset+4]))
	height := int(binary.BigEndian.Uint32(b[offset+4 : offset+8]))
	return width, height, nil
}

// 获取 BMP 图片的宽高
func GetBmpWidthHeight(b []byte) (w, h int, err error) {
	if string(b[:2]) != "BM" {
		return 0, 0, errors.New("GetBmpWidthHeight: not a bmp file")
	}
	width := int(binary.LittleEndian.Uint32(b[18:22]))
	height := int(int32(binary.LittleEndian.Uint32(b[22:26])))
	if height < 0 {
		height = -height
	}
	return width, height, nil
}
