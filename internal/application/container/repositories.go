package container

import (
	"context"

	"fms-project/internal/domain/repository"
	"fms-project/internal/domain/transaction"
	"fms-project/internal/infrastructure/config"
	"fms-project/internal/infrastructure/logger"
	inmemory "fms-project/internal/infrastructure/storage/in-memory"
	"fms-project/internal/infrastructure/storage/postgres"
	"fms-project/internal/infrastructure/storage/postgres/repository/category"
	"fms-project/internal/infrastructure/storage/postgres/repository/organisation"
	"fms-project/internal/infrastructure/storage/postgres/repository/purchase"
)

type RepositoriesContainer struct {
	Categories    repository.CategoryRepository
	Purchases     repository.PurchaseRepository
	Organisations repository.OrganisationRepository
	UserState     repository.UserStateRepository
	TxManager     transaction.TxManager
}

type RepositoriesContainerConfig struct {
	Config   *config.Config
	Logger   logger.Logger
	Postgres *postgres.Client
}

func NewRepositoriesContainer(ctx context.Context, cfg *RepositoriesContainerConfig) *RepositoriesContainer {
	txManager := postgres.NewTxManager(cfg.Postgres)

	categoryRepository := category.NewCategoryRepository(&category.CategoryRepositoryConfig{
		Logger: cfg.Logger,
		Client: cfg.Postgres,
	})

	purchaseRepository := purchase.NewPurchaseRepository(&purchase.PurchaseRepositoryConfig{
		Logger: cfg.Logger,
		Client: cfg.Postgres,
	})

	organisationRepository := organisation.NewOrganisationRepository(&organisation.OrganisationRepositoryConfig{
		Logger: cfg.Logger,
		Client: cfg.Postgres,
	})

	userStateRepository := inmemory.NewUserStateRepository()

	return &RepositoriesContainer{
		Categories:    categoryRepository,
		Purchases:     purchaseRepository,
		Organisations: organisationRepository,
		UserState:     userStateRepository,
		TxManager:     txManager,
	}
}
