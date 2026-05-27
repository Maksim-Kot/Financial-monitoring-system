package container

import (
	"context"

	"fms-project/internal/application/service"
	"fms-project/internal/infrastructure/config"
	"fms-project/internal/infrastructure/logger"
)

type ServicesContainer struct {
	ImagePreprocessor *service.ImagePreprocessorService
	CategoryMatcher   *service.CategoryMatcherService
}

type ServicesContainerConfig struct {
	Config       *config.Config
	Logger       logger.Logger
	Repositories *RepositoriesContainer
	Gateways     *GatewaysContainer
}

func NewServicesContainer(ctx context.Context, cfg *ServicesContainerConfig) *ServicesContainer {
	imagePreprocessorService := service.NewImagePreprocessorService(&service.ImagePreprocessorServiceConfig{})
	categoryMatcherService := service.NewCategoryMatcherService(&service.CategoryMatcherServiceConfig{})

	return &ServicesContainer{
		ImagePreprocessor: imagePreprocessorService,
		CategoryMatcher:   categoryMatcherService,
	}
}
