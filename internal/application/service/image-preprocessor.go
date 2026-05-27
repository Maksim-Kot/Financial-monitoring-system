package service

import "fms-project/internal/infrastructure/adapter"

type ImagePreprocessorServiceConfig struct{}

type ImagePreprocessorService struct {
	processor *adapter.ImagePreprocessor
}

func NewImagePreprocessorService(cfg *ImagePreprocessorServiceConfig) *ImagePreprocessorService {
	processor := adapter.NewImagePreprocessor()

	return &ImagePreprocessorService{
		processor: processor,
	}
}

func (s *ImagePreprocessorService) Process(payload []byte) ([]byte, error) {
	return s.processor.Process(payload)
}
