package usecase

import (
	"context"

	"fms-project/internal/application/shared"
	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/repository"
	"fms-project/internal/domain/valueobject"
	"fms-project/internal/infrastructure/logger"
)

type ListScenarioUseCaseRequest struct {
	UserID int64
	Limit  int
	Offset int
}

type ListScenarioUseCaseResponse struct {
	Purchases []entity.Purchase
	Total     int
}

type ListScenarioUseCaseConfig struct {
	Logger             logger.Logger
	PurchaseRepository repository.PurchaseRepository
}

type ListScenarioUseCase struct {
	logger             logger.Logger
	purchaseRepository repository.PurchaseRepository
}

func NewListScenarioUseCase(cfg *ListScenarioUseCaseConfig) shared.UseCase[ListScenarioUseCaseRequest, ListScenarioUseCaseResponse] {
	return &ListScenarioUseCase{
		logger:             cfg.Logger.With("layer", "usecase", "usecase", "ListScenario"),
		purchaseRepository: cfg.PurchaseRepository,
	}
}

func (uc *ListScenarioUseCase) Execute(ctx context.Context, in ListScenarioUseCaseRequest) (ListScenarioUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID)

	listOut, err := uc.purchaseRepository.GetByUserID(ctx, repository.PurchaseRepositoryGetByUserIDIn{
		UserID: in.UserID,
		Limit:  in.Limit,
		Offset: in.Offset,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get purchases list", "error", err)
		return ListScenarioUseCaseResponse{}, err
	}

	if !listOut.Exists() {
		return ListScenarioUseCaseResponse{Purchases: nil, Total: listOut.Total}, nil
	}

	ids := make([]valueobject.UUID, 0, len(listOut.Purchases))
	for _, item := range listOut.Purchases {
		ids = append(ids, item.ID)
	}

	purchasesOut, err := uc.purchaseRepository.GetByIDs(ctx, repository.PurchaseRepositoryGetByIDsIn{
		IDs: ids,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get purchases by IDs", "error", err)
		return ListScenarioUseCaseResponse{}, err
	}

	return ListScenarioUseCaseResponse{
		Purchases: purchasesOut.Purchases,
		Total:     listOut.Total,
	}, nil
}
