package usecase

import (
	"context"

	"fms-project/internal/application/shared"
	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/repository"
	"fms-project/internal/infrastructure/logger"
)

type GetPurchasesByPeriodUseCaseRequest struct {
	UserID int64
	Year   int
	Month  int
	Limit  int
	Offset int
}

type GetPurchasesByPeriodUseCaseResponse struct {
	Purchases []entity.Purchase
	Total     int
}

type GetPurchasesByPeriodUseCaseConfig struct {
	Logger             logger.Logger
	PurchaseRepository repository.PurchaseRepository
}

type GetPurchasesByPeriodUseCase struct {
	logger             logger.Logger
	purchaseRepository repository.PurchaseRepository
}

func NewGetPurchasesByPeriodUseCase(cfg *GetPurchasesByPeriodUseCaseConfig) shared.UseCase[GetPurchasesByPeriodUseCaseRequest, GetPurchasesByPeriodUseCaseResponse] {
	return &GetPurchasesByPeriodUseCase{
		logger:             cfg.Logger.With("layer", "usecase", "usecase", "GetPurchasesByPeriod"),
		purchaseRepository: cfg.PurchaseRepository,
	}
}

func (uc *GetPurchasesByPeriodUseCase) Execute(ctx context.Context, in GetPurchasesByPeriodUseCaseRequest) (GetPurchasesByPeriodUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID, "year", in.Year, "month", in.Month)

	purchasesOut, err := uc.purchaseRepository.GetByUserIDAndPeriod(ctx, repository.PurchaseRepositoryGetByUserIDAndPeriodIn{
		UserID: in.UserID,
		Year:   in.Year,
		Month:  in.Month,
		Limit:  in.Limit,
		Offset: in.Offset,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get purchases by period", "error", err)
		return GetPurchasesByPeriodUseCaseResponse{}, err
	}

	return GetPurchasesByPeriodUseCaseResponse{
		Purchases: purchasesOut.Purchases,
		Total:     purchasesOut.Total,
	}, nil
}
