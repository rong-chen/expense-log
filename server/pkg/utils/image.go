package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // 注册 PNG 解码器
	"io"
)

// CompressImage 压缩图片质量，输出为 JPEG
// maxWidth: 最大宽度（等比缩放），quality: JPEG 质量 (1-100)
func CompressImage(src io.Reader, maxWidth int, quality int) ([]byte, error) {
	img, _, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	// 等比缩放（仅缩小，不放大）
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	newW := origW
	newH := origH
	if origW > maxWidth {
		newW = maxWidth
		newH = origH * maxWidth / origW
	}

	// 使用简单的最近邻缩放（避免引入额外依赖）
	resized := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX := x * origW / newW
			srcY := y * origH / newH
			resized.Set(x, y, img.At(srcX+bounds.Min.X, srcY+bounds.Min.Y))
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("encode jpeg: %w", err)
	}

	return buf.Bytes(), nil
}
