package usecase

import (
	"context"
	"strconv"
	"strings"

	"fms-project/internal/application/shared"
	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/repository"
	domainError "fms-project/internal/domain/shared/domain-error"
	"fms-project/internal/infrastructure/logger"
)

type ManualAddItemUseCaseRequest struct {
	UserID int64
	Item   string
}

type ManualAddItemUseCaseResponse struct {
	Item entity.DraftItem
}

type ManualAddItemUseCaseConfig struct {
	Logger          logger.Logger
	StateRepository repository.UserStateRepository
}

type ManualAddItemUseCase struct {
	logger          logger.Logger
	stateRepository repository.UserStateRepository
}

func NewManualAddItemUseCase(cfg *ManualAddItemUseCaseConfig) shared.UseCase[ManualAddItemUseCaseRequest, ManualAddItemUseCaseResponse] {
	return &ManualAddItemUseCase{
		logger:          cfg.Logger.With("layer", "usecase", "usecase", "ManualAddItem"),
		stateRepository: cfg.StateRepository,
	}
}

func (uc *ManualAddItemUseCase) Execute(ctx context.Context, in ManualAddItemUseCaseRequest) (ManualAddItemUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID)

	item, err := uc.parseItem(in.Item)
	if err != nil {
		logger.DebugContext(ctx, "failed to parse manual item", "error", err, "input", in.Item)
		return ManualAddItemUseCaseResponse{}, domainError.New(domainError.StatusValidation, err.Error())
	}

	stateOut, err := uc.stateRepository.Get(ctx, repository.UserStateRepositoryGetIn{UserID: in.UserID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get state", "error", err)
		return ManualAddItemUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to get state")
	}

	var state *entity.UserState
	if stateOut.Exists() {
		state = stateOut.UserState
	} else {
		state = &entity.UserState{
			UserID:     in.UserID,
			DraftItems: []entity.DraftItem{},
			SourceType: entity.SourceTypeManual,
		}
	}

	state.DraftItems = append(state.DraftItems, item)

	if _, err := uc.stateRepository.Save(ctx, repository.UserStateRepositorySaveIn{UserState: state}); err != nil {
		logger.ErrorContext(ctx, "failed to save state", "error", err)
		return ManualAddItemUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to save state")
	}

	logger.DebugContext(ctx, "added manual item", "name", item.Name, "quantity", item.Quantity, "unitPrice", item.UnitPrice)

	return ManualAddItemUseCaseResponse{Item: item}, nil
}

func (uc *ManualAddItemUseCase) parseItem(input string) (entity.DraftItem, error) {
	parts := strings.Split(input, ";")
	if len(parts) != 3 {
		return entity.DraftItem{}, domainError.New(domainError.StatusValidation, "invalid input format")
	}

	name := strings.TrimSpace(parts[0])
	if name == "" {
		return entity.DraftItem{}, domainError.New(domainError.StatusValidation, "name cannot be empty")
	}

	quantityStr := strings.TrimSpace(parts[1])
	quantity, err := strconv.ParseFloat(quantityStr, 64)
	if err != nil || quantity <= 0 {
		return entity.DraftItem{}, domainError.New(domainError.StatusValidation, "quantity must be a positive number")
	}

	unitPriceStr := strings.TrimSpace(parts[2])
	unitPrice, err := strconv.ParseFloat(unitPriceStr, 64)
	if err != nil || unitPrice < 0 {
		return entity.DraftItem{}, domainError.New(domainError.StatusValidation, "price must be a positive number")
	}

	return entity.DraftItem{
		Name:      name,
		Quantity:  quantity,
		UnitPrice: unitPrice,
		Category:  entity.Category{},
	}, nil
}
