package bootstrap

import (
	"context"

	"fms-project/internal/application/shared"
	"fms-project/internal/application/usecase"
	"fms-project/internal/infrastructure/logger"
)

type Bootstrap interface {
	Run(ctx context.Context) error
}

type bootstrap struct {
	ensureCategories shared.UseCase[usecase.EnsureCategoriesUseCaseRequest, usecase.EnsureCategoriesUseCaseResponse]
	logger           logger.Logger
}

type BootstrapConfig struct {
	EnsureCategories shared.UseCase[usecase.EnsureCategoriesUseCaseRequest, usecase.EnsureCategoriesUseCaseResponse]
	Logger           logger.Logger
}

func New(cfg *BootstrapConfig) Bootstrap {
	return &bootstrap{
		ensureCategories: cfg.EnsureCategories,
		logger:           cfg.Logger,
	}
}

func (b *bootstrap) Run(ctx context.Context) error {
	_, err := b.ensureCategories.Execute(ctx, usecase.EnsureCategoriesUseCaseRequest{})
	if err != nil {
		return err
	}

	b.logger.InfoContext(ctx, "сategories were bootstrapped")
	return nil
}
