package container

import (
	"fms-project/internal/infrastructure/config"
	"fms-project/internal/infrastructure/logger"
)

type UseCasesContainer struct {
}

type UseCasesContainerConfig struct {
	Config       *config.Config
	Logger       logger.Logger
	Services     *ServicesContainer
	Repositories *RepositoriesContainer
	Gateways     *GatewaysContainer
}

func NewUseCasesContainer(cfg *UseCasesContainerConfig) *UseCasesContainer {
	return &UseCasesContainer{}
}
