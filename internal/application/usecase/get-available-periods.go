package usecase

import (
	"context"

	"fms-project/internal/application/shared"
	"fms-project/internal/domain/repository"
	"fms-project/internal/infrastructure/logger"
)

type GetAvailablePeriodsUseCaseRequest struct {
	UserID int64
}

type GetAvailablePeriodsUseCaseResponse struct {
	Periods []repository.YearWithMonths
}

type GetAvailablePeriodsUseCaseConfig struct {
	Logger             logger.Logger
	PurchaseRepository repository.PurchaseRepository
}

type GetAvailablePeriodsUseCase struct {
	logger             logger.Logger
	purchaseRepository repository.PurchaseRepository
}

func NewGetAvailablePeriodsUseCase(cfg *GetAvailablePeriodsUseCaseConfig) shared.UseCase[GetAvailablePeriodsUseCaseRequest, GetAvailablePeriodsUseCaseResponse] {
	return &GetAvailablePeriodsUseCase{
		logger:             cfg.Logger.With("layer", "usecase", "usecase", "GetAvailablePeriods"),
		purchaseRepository: cfg.PurchaseRepository,
	}
}

func (uc *GetAvailablePeriodsUseCase) Execute(ctx context.Context, in GetAvailablePeriodsUseCaseRequest) (GetAvailablePeriodsUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID)

	periodsOut, err := uc.purchaseRepository.GetAvailablePeriods(ctx, repository.PurchaseRepositoryGetAvailablePeriodsIn{
		UserID: in.UserID,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get available periods", "error", err)
		return GetAvailablePeriodsUseCaseResponse{}, err
	}

	return GetAvailablePeriodsUseCaseResponse{
		Periods: periodsOut.Periods,
	}, nil
}
