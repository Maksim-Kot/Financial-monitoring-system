package container

import (
	"context"

	"fms-project/internal/application/service"
	"fms-project/internal/infrastructure/config"
	"fms-project/internal/infrastructure/logger"
)

type ServicesContainer struct {
	ImagePreprocessor *service.ImagePreprocessorService
}

type ServicesContainerConfig struct {
	Config       *config.Config
	Logger       logger.Logger
	Repositories *RepositoriesContainer
	Gateways     *GatewaysContainer
}

func NewServicesContainer(ctx context.Context, cfg *ServicesContainerConfig) *ServicesContainer {
	imagePreprocessorService := service.NewImagePreprocessorService(&service.ImagePreprocessorServiceConfig{})

	return &ServicesContainer{
		ImagePreprocessor: imagePreprocessorService,
	}
}
