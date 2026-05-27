package usecase

import (
	"context"

	"fms-project/internal/application/service"
	"fms-project/internal/application/shared"
	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/gateway"
	"fms-project/internal/domain/repository"
	domainError "fms-project/internal/domain/shared/domain-error"
	"fms-project/internal/infrastructure/logger"
)

type ClassifyCategoryUseCaseRequest struct {
	UserID int64
}

type ClassifyCategoryUseCaseResponse struct {
}

type ClassifyCategoryUseCaseConfig struct {
	Logger                    logger.Logger
	StateRepository           repository.UserStateRepository
	CategoryRepository        repository.CategoryRepository
	CategoryMatcher           *service.CategoryMatcherService
	CategoryClassifierGateway gateway.CategoryClassifierGateway
}

type ClassifyCategoryUseCase struct {
	logger                    logger.Logger
	stateRepository           repository.UserStateRepository
	categoryRepository        repository.CategoryRepository
	categoryMatcher           *service.CategoryMatcherService
	categoryClassifierGateway gateway.CategoryClassifierGateway
}

func NewClassifyCategoryUseCase(cfg *ClassifyCategoryUseCaseConfig) shared.UseCase[ClassifyCategoryUseCaseRequest, ClassifyCategoryUseCaseResponse] {
	return &ClassifyCategoryUseCase{
		logger:                    cfg.Logger.With("layer", "usecase", "usecase", "ClassifyCategory"),
		stateRepository:           cfg.StateRepository,
		categoryRepository:        cfg.CategoryRepository,
		categoryMatcher:           cfg.CategoryMatcher,
		categoryClassifierGateway: cfg.CategoryClassifierGateway,
	}
}

func (uc *ClassifyCategoryUseCase) Execute(ctx context.Context, in ClassifyCategoryUseCaseRequest) (ClassifyCategoryUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID)

	stateOut, err := uc.stateRepository.Get(ctx, repository.UserStateRepositoryGetIn{UserID: in.UserID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get user state", "error", err)
		return ClassifyCategoryUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to get user state")
	}
	if !stateOut.Exists() {
		logger.ErrorContext(ctx, "user state not found")
		return ClassifyCategoryUseCaseResponse{}, domainError.New(domainError.StatusNotFound, "user state not found")
	}

	items := stateOut.UserState.DraftItems
	if len(items) == 0 {
		logger.InfoContext(ctx, "no items to classify")
		return ClassifyCategoryUseCaseResponse{}, nil
	}

	logger.DebugContext(ctx, "loaded items for classification", "count", len(items))

	categoriesOut, err := uc.categoryRepository.List(ctx, repository.CategoryRepositoryListIn{})
	if err != nil {
		logger.ErrorContext(ctx, "failed to list categories", "error", err)
		return ClassifyCategoryUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to get categories")
	}
	if !categoriesOut.Exists() {
		logger.ErrorContext(ctx, "no categories found")
		return ClassifyCategoryUseCaseResponse{}, domainError.New(domainError.StatusNotFound, "no categories found")
	}
	categories := categoriesOut.Categories

	logger.DebugContext(ctx, "loaded categories", "count", len(categories))

	items = uc.categoryMatcher.Match(items, categories)

	itemsWithoutCategory := make([]entity.DraftItem, 0)
	itemsWithCategory := make([]entity.DraftItem, 0)
	for _, item := range items {
		if item.Category.Name == "" {
			itemsWithoutCategory = append(itemsWithoutCategory, item)
		} else {
			itemsWithCategory = append(itemsWithCategory, item)
		}
	}

	logger.DebugContext(ctx, "items after dictionary match", "matched", len(itemsWithCategory), "unmatched", len(itemsWithoutCategory))

	if len(itemsWithoutCategory) > 0 {
		classifyOut, err := uc.categoryClassifierGateway.ClassifyCategories(ctx, gateway.CategoryClassifierGatewayIn{
			Items:      itemsWithoutCategory,
			Categories: categories,
		})
		if err != nil {
			logger.ErrorContext(ctx, "failed to classify categories via provider", "error", err)
		}

		itemsWithCategory = append(itemsWithCategory, classifyOut.Items...)
	}

	stateOut.UserState.DraftItems = itemsWithCategory
	if _, err := uc.stateRepository.Save(ctx, repository.UserStateRepositorySaveIn{UserState: stateOut.UserState}); err != nil {
		logger.ErrorContext(ctx, "failed to save state with categories", "error", err)
		return ClassifyCategoryUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to save state")
	}

	logger.InfoContext(ctx, "classification completed", "items", len(itemsWithCategory))

	return ClassifyCategoryUseCaseResponse{}, nil
}
