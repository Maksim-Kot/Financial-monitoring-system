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
	"fms-project/internal/domain/valueobject"
	"fms-project/internal/infrastructure/logger"
)

type UpdateExpenseUseCaseRequest struct {
	UserID     int64
	PurchaseID valueobject.UUID
	ExpenseID  valueobject.UUID
	Name       string
	Quantity   float64
	UnitPrice  float64 // in decimal format (e.g., 1.50 for 1.50 BYN)
}

type UpdateExpenseUseCaseResponse struct {
	Expense entity.Expense
}

type UpdateExpenseUseCaseConfig struct {
	Logger             logger.Logger
	PurchaseRepository repository.PurchaseRepository
}

type UpdateExpenseUseCase struct {
	logger             logger.Logger
	purchaseRepository repository.PurchaseRepository
}

func NewUpdateExpenseUseCase(cfg *UpdateExpenseUseCaseConfig) shared.UseCase[UpdateExpenseUseCaseRequest, UpdateExpenseUseCaseResponse] {
	return &UpdateExpenseUseCase{
		logger:             cfg.Logger.With("layer", "usecase", "usecase", "UpdateExpense"),
		purchaseRepository: cfg.PurchaseRepository,
	}
}

func (uc *UpdateExpenseUseCase) Execute(ctx context.Context, in UpdateExpenseUseCaseRequest) (UpdateExpenseUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID, "purchaseID", in.PurchaseID.String(), "expenseID", in.ExpenseID.String())

	// Validate input
	if err := uc.validateInput(in); err != nil {
		logger.DebugContext(ctx, "invalid input", "error", err)
		return UpdateExpenseUseCaseResponse{}, domainError.New(domainError.StatusValidation, err.Error())
	}

	// Get the purchase with the expense
	purchaseOut, err := uc.purchaseRepository.GetByIDs(ctx, repository.PurchaseRepositoryGetByIDsIn{
		IDs: []valueobject.UUID{in.PurchaseID},
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get purchase", "error", err)
		return UpdateExpenseUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to get purchase")
	}
	if !purchaseOut.Exists() || len(purchaseOut.Purchases) == 0 {
		return UpdateExpenseUseCaseResponse{}, domainError.New(domainError.StatusNotFound, "purchase not found")
	}

	purchase := purchaseOut.Purchases[0]

	// Verify the purchase belongs to the user
	if purchase.UserID != in.UserID {
		return UpdateExpenseUseCaseResponse{}, domainError.New(domainError.StatusForbidden, "purchase does not belong to user")
	}

	// Find and update the expense
	var updatedExpense entity.Expense
	found := false
	expenses := purchase.Expenses()

	for i, exp := range expenses {
		if exp.ID == in.ExpenseID {
			found = true

			// Create new unit price money amount
			unitMinor := int64(math.Round(in.UnitPrice * 100))
			unitPrice, err := valueobject.NewMoneyAmountFromInt64(unitMinor, 2, valueobject.MoneyAmountCurrencyBYN, nil)
			if err != nil {
				logger.ErrorContext(ctx, "failed to create unit price", "error", err)
				return UpdateExpenseUseCaseResponse{}, domainError.New(domainError.StatusValidation, "invalid unit price")
			}

			// Update the expense
			expenses[i].SetName(strings.TrimSpace(in.Name))
			expenses[i].SetQuantity(in.Quantity)
			expenses[i].SetUnitPrice(unitPrice)
			// Category remains unchanged

			updatedExpense = expenses[i]
			break
		}
	}

	if !found {
		return UpdateExpenseUseCaseResponse{}, domainError.New(domainError.StatusNotFound, "expense not found in purchase")
	}

	// Rebuild the purchase with updated expenses
	updatedPurchase := rebuildPurchase(purchase, expenses)

	_, err = uc.purchaseRepository.Save(ctx, repository.PurchaseRepositorySaveIn{
		Purchase: updatedPurchase,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to save purchase with updated expense", "error", err)
		return UpdateExpenseUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to save updated expense")
	}

	return UpdateExpenseUseCaseResponse{
		Expense: updatedExpense,
	}, nil
}

func (uc *UpdateExpenseUseCase) validateInput(in UpdateExpenseUseCaseRequest) error {
	if strings.TrimSpace(in.Name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if in.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if in.UnitPrice < 0 {
		return fmt.Errorf("unit price cannot be negative")
	}
	return nil
}

func rebuildPurchase(original entity.Purchase, updatedExpenses []entity.Expense) entity.Purchase {
	newPurchase := entity.NewPurchase(original.UserID, original.SourceType)

	newPurchase.ID = original.ID
	newPurchase.SetPurchaseDate(original.PurchaseDate)
	newPurchase.SetOrganisation(original.OrganisationName)
	newPurchase.SetDescription(original.Description)

	for _, exp := range updatedExpenses {
		newPurchase.AddExpense(exp)
	}

	return newPurchase
}
