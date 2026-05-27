package usecase

import (
	"context"

	"fms-project/internal/application/shared"
	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/repository"
	"fms-project/internal/domain/transaction"
	"fms-project/internal/infrastructure/logger"
)

type EnsureCategoriesUseCaseRequest struct {
}

type EnsureCategoriesUseCaseResponse struct {
}

type EnsureCategoriesUseCaseConfig struct {
	Logger             logger.Logger
	CategoryRepository repository.CategoryRepository
	TxManager          transaction.TxManager
}

type EnsureCategoriesUseCase struct {
	logger             logger.Logger
	categoryRepository repository.CategoryRepository
	txManager          transaction.TxManager
}

func NewEnsureCategoriesUseCase(cfg *EnsureCategoriesUseCaseConfig) shared.UseCase[EnsureCategoriesUseCaseRequest, EnsureCategoriesUseCaseResponse] {
	return &EnsureCategoriesUseCase{
		logger:             cfg.Logger.With("layer", "usecase", "usecase", "EnsureCategoriesUseCase"),
		categoryRepository: cfg.CategoryRepository,
		txManager:          cfg.TxManager,
	}
}

func (uc *EnsureCategoriesUseCase) Execute(ctx context.Context, in EnsureCategoriesUseCaseRequest) (EnsureCategoriesUseCaseResponse, error) {
	categories, err := uc.categoryRepository.List(ctx, repository.CategoryRepositoryListIn{})
	if err != nil {
		uc.logger.ErrorContext(ctx, "failed to list categories", "error", err)
		return EnsureCategoriesUseCaseResponse{}, err
	}

	if categories.Exists() {
		uc.logger.InfoContext(ctx, "categories already exist")
		return EnsureCategoriesUseCaseResponse{}, nil
	}

	uc.logger.InfoContext(ctx, "no categories found, creating default categories")

	err = uc.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		for _, seed := range entity.DefaultCategorySeeds {
			category := entity.NewCategory(seed.Name, seed.Icon)
			_, err := uc.categoryRepository.Create(txCtx, repository.CategoryRepositoryCreateIn{Category: category})
			if err != nil {
				uc.logger.ErrorContext(txCtx, "failed to create category", "error", err, "categoryName", category.Name, "categoryIcon", category.Icon)
				return err
			}
		}
		return nil
	})

	if err != nil {
		uc.logger.ErrorContext(ctx, "failed to create categories in transaction", "error", err)
		return EnsureCategoriesUseCaseResponse{}, err
	}

	uc.logger.InfoContext(ctx, "default categories created successfully")
	return EnsureCategoriesUseCaseResponse{}, nil
}
