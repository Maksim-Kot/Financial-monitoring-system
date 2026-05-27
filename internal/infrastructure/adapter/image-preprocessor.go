package adapter

import (
	"bytes"

	"github.com/disintegration/imaging"
)

type ImagePreprocessor struct{}

func NewImagePreprocessor() *ImagePreprocessor {
	return &ImagePreprocessor{}
}

func (p *ImagePreprocessor) Process(payload []byte) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(payload), imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}

	processed := imaging.Grayscale(img)
	processed = imaging.AdjustContrast(processed, 45)
	processed = imaging.Sharpen(processed, 10)
	processed = imaging.AdjustBrightness(processed, 0)
	processed = imaging.AdjustSaturation(processed, 0)

	var out bytes.Buffer
	if err := imaging.Encode(&out, processed, imaging.JPEG); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
