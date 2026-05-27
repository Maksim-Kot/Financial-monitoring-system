package container

import (
	"context"

	"fms-project/internal/infrastructure/config"
	"fms-project/internal/infrastructure/logger"
	"fms-project/internal/infrastructure/storage/postgres"
)

type Container struct {
	Gateways     *GatewaysContainer
	Repositories *RepositoriesContainer
	Services     *ServicesContainer
	Usecases     *UseCasesContainer
}

type ContainerConfig struct {
	Config   *config.Config
	Logger   logger.Logger
	Postgres *postgres.Client
}

func NewContainer(ctx context.Context, cfg *ContainerConfig) *Container {
	gateways := NewGatewaysContainer(&GatewaysContainerConfig{
		Config: cfg.Config,
		Logger: cfg.Logger,
	})

	repositories := NewRepositoriesContainer(ctx, &RepositoriesContainerConfig{
		Logger:   cfg.Logger,
		Postgres: cfg.Postgres,
	})

	services := NewServicesContainer(ctx, &ServicesContainerConfig{
		Config:       cfg.Config,
		Logger:       cfg.Logger,
		Repositories: repositories,
		Gateways:     gateways,
	})

	usecases := NewUseCasesContainer(&UseCasesContainerConfig{
		Config:       cfg.Config,
		Logger:       cfg.Logger,
		Repositories: repositories,
		Gateways:     gateways,
		Services:     services,
	})

	return &Container{
		Gateways:     gateways,
		Repositories: repositories,
		Services:     services,
		Usecases:     usecases,
	}
}
