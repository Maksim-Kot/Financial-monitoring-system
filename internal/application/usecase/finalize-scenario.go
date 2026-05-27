package usecase

import (
	"context"
	"fmt"
	"math"
	"strings"

	"fms-project/internal/application/shared"
	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/repository"
	domainError "fms-project/internal/domain/shared/domain-error"
	"fms-project/internal/domain/transaction"
	"fms-project/internal/domain/valueobject"
	"fms-project/internal/infrastructure/logger"
)

type FinalizeScenarioUseCaseRequest struct {
	UserID int64
}

type FinalizeScenarioUseCaseResponse struct {
}

type FinalizeScenarioUseCaseConfig struct {
	Logger             logger.Logger
	StateRepository    repository.UserStateRepository
	PurchaseRepository repository.PurchaseRepository
	TxManager          transaction.TxManager
}

type FinalizeScenarioUseCase struct {
	logger             logger.Logger
	stateRepository    repository.UserStateRepository
	purchaseRepository repository.PurchaseRepository
	txManager          transaction.TxManager
}

func NewFinalizeScenarioUseCase(cfg *FinalizeScenarioUseCaseConfig) shared.UseCase[FinalizeScenarioUseCaseRequest, FinalizeScenarioUseCaseResponse] {
	return &FinalizeScenarioUseCase{
		logger:             cfg.Logger.With("layer", "usecase", "usecase", "FinalizeScenario"),
		stateRepository:    cfg.StateRepository,
		purchaseRepository: cfg.PurchaseRepository,
		txManager:          cfg.TxManager,
	}
}

func (uc *FinalizeScenarioUseCase) Execute(ctx context.Context, in FinalizeScenarioUseCaseRequest) (FinalizeScenarioUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID)

	stateOut, err := uc.stateRepository.Get(ctx, repository.UserStateRepositoryGetIn{UserID: in.UserID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get state", "error", err)
		return FinalizeScenarioUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to get state")
	}
	if !stateOut.Exists() {
		return FinalizeScenarioUseCaseResponse{}, domainError.New(domainError.StatusNotFound, "state not found")
	}

	state := stateOut.UserState
	if state.UserID != in.UserID {
		return FinalizeScenarioUseCaseResponse{}, domainError.New(domainError.StatusConflict, "user ID mismatch")
	}

	if state.SourceType == "" {
		return FinalizeScenarioUseCaseResponse{}, domainError.New(domainError.StatusConflict, "source type is empty")
	}

	if state.PurchaseDate.IsZero() {
		return FinalizeScenarioUseCaseResponse{}, domainError.New(domainError.StatusConflict, "purchase date is empty")
	}

	if err := uc.savePurchase(ctx, state); err != nil {
		logger.ErrorContext(ctx, "failed to save purchase", "error", err)
		return FinalizeScenarioUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to save purchase")
	}

	if _, err := uc.stateRepository.Delete(ctx, repository.UserStateRepositoryDeleteIn{UserID: in.UserID}); err != nil {
		logger.ErrorContext(ctx, "failed to delete state", "error", err)
		return FinalizeScenarioUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to delete state")
	}

	return FinalizeScenarioUseCaseResponse{}, nil
}

func (uc *FinalizeScenarioUseCase) savePurchase(ctx context.Context, state *entity.UserState) error {
	if len(state.DraftItems) == 0 {
		return fmt.Errorf("draft items are empty")
	}
	if state.PurchaseDate.IsZero() {
		return fmt.Errorf("purchase date is empty")
	}
	if state.Organisation == "" {
		return fmt.Errorf("organisation is empty")
	}

	return uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		sourceType := state.SourceType
		if sourceType == "" {
			sourceType = entity.SourceTypeText
		}

		purchase := entity.NewPurchase(state.UserID, sourceType)
		purchase.SetPurchaseDate(state.PurchaseDate)
		purchase.SetOrganisation(state.Organisation)
		purchase.SetDescription(buildDescription(state.DraftItems))

		for _, item := range state.DraftItems {
			expense, err := buildExpense(purchase.ID, item)
			if err != nil {
				return err
			}
			purchase.AddExpense(expense)
		}

		_, err := uc.purchaseRepository.Save(ctx, repository.PurchaseRepositorySaveIn{
			Purchase: purchase,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func buildExpense(purchaseID valueobject.UUID, item entity.DraftItem) (entity.Expense, error) {
	name := strings.TrimSpace(item.Name)
	if name == "" {
		name = "unknown"
	}

	quantity := item.Quantity
	if quantity <= 0 {
		quantity = 1
	}

	unitMinor := int64(math.Round(item.UnitPrice * 100))
	if unitMinor < 0 {
		return entity.Expense{}, fmt.Errorf("negative unit price")
	}

	unitPrice, err := valueobject.NewMoneyAmountFromInt64(unitMinor, 2, valueobject.MoneyAmountCurrencyBYN, nil)
	if err != nil {
		return entity.Expense{}, err
	}

	expense := entity.NewExpense(purchaseID)
	expense.SetName(name)
	expense.SetQuantity(quantity)
	expense.SetUnitPrice(unitPrice)
	expense.SetCategory(item.Category)

	return expense, nil
}

func buildDescription(items []entity.DraftItem) string {
	if len(items) == 0 {
		return ""
	}

	names := make([]string, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		names = append(names, name)
		if len(names) == 3 {
			break
		}
	}
	if len(names) == 0 {
		return ""
	}
	if len(items) > 3 {
		return strings.Join(names, ", ") + ", ..."
	}
	return strings.Join(names, ", ")
}
