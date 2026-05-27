package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fms-project/internal/application/shared"
	"fms-project/internal/domain/repository"
	domainError "fms-project/internal/domain/shared/domain-error"
	"fms-project/internal/infrastructure/logger"
)

type AddPurchaseDateUseCaseRequest struct {
	UserID int64
	Date   string
}

type AddPurchaseDateUseCaseResponse struct {
}

type AddPurchaseDateUseCaseConfig struct {
	Logger          logger.Logger
	StateRepository repository.UserStateRepository
}

type AddPurchaseDateUseCase struct {
	logger          logger.Logger
	stateRepository repository.UserStateRepository
}

func NewAddPurchaseDateUseCase(cfg *AddPurchaseDateUseCaseConfig) shared.UseCase[AddPurchaseDateUseCaseRequest, AddPurchaseDateUseCaseResponse] {
	return &AddPurchaseDateUseCase{
		logger:          cfg.Logger.With("layer", "usecase", "usecase", "AddPurchaseDate"),
		stateRepository: cfg.StateRepository,
	}
}

func (uc *AddPurchaseDateUseCase) Execute(ctx context.Context, in AddPurchaseDateUseCaseRequest) (AddPurchaseDateUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID)

	stateOut, err := uc.stateRepository.Get(ctx, repository.UserStateRepositoryGetIn{UserID: in.UserID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get state", "error", err)
		return AddPurchaseDateUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to get state")
	}
	if !stateOut.Exists() {
		return AddPurchaseDateUseCaseResponse{}, domainError.New(domainError.StatusNotFound, "state not found")
	}

	state := stateOut.UserState
	if state.UserID != in.UserID {
		return AddPurchaseDateUseCaseResponse{}, domainError.New(domainError.StatusConflict, "user ID mismatch")
	}

	if state.SourceType == "" {
		return AddPurchaseDateUseCaseResponse{}, domainError.New(domainError.StatusConflict, "source type is empty")
	}

	purchaseDate, err := parsePurchaseDate(in.Date)
	if err != nil {
		return AddPurchaseDateUseCaseResponse{}, domainError.New(domainError.StatusValidation, "failed to parse purchase date")
	}
	state.PurchaseDate = purchaseDate
	if _, err := uc.stateRepository.Save(ctx, repository.UserStateRepositorySaveIn{UserState: state}); err != nil {
		logger.ErrorContext(ctx, "failed to save state", "error", err)
		return AddPurchaseDateUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to save state")
	}

	return AddPurchaseDateUseCaseResponse{}, nil
}

func parsePurchaseDate(raw string) (time.Time, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	layouts := []string{
		"02.01.2006",
		"2006-01-02",
		"02/01/2006",
		"02-01-2006",
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, text)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date format")
}
