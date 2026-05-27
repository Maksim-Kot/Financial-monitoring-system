package container

import (
	"context"

	"fms-project/internal/application/service"
	"fms-project/internal/infrastructure/config"
	"fms-project/internal/infrastructure/logger"
)

type ServicesContainer struct {
	ImagePreprocessor   *service.ImagePreprocessorService
	CategoryMatcher     *service.CategoryMatcherService
	AnalyticsCalculator service.AnalyticsCalculator
	AnomalyDetector     service.AnomalyDetector
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

	analyticsCalculatorService := service.NewAnalyticsCalculatorService()
	anomalyDetectorService := service.NewAnomalyDetectorService(2.0) // threshold for anomaly detection

	return &ServicesContainer{
		ImagePreprocessor:   imagePreprocessorService,
		CategoryMatcher:     categoryMatcherService,
		AnalyticsCalculator: analyticsCalculatorService,
		AnomalyDetector:     anomalyDetectorService,
	}
}
