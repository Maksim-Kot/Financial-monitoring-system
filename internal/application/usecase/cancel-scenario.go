package usecase

import (
	"context"

	"fms-project/internal/application/shared"
	"fms-project/internal/domain/repository"
	domainError "fms-project/internal/domain/shared/domain-error"
	"fms-project/internal/infrastructure/logger"
)

type CancelScenarioUseCaseRequest struct {
	UserID int64
}

type CancelScenarioUseCaseResponse struct {
}

type CancelScenarioUseCaseConfig struct {
	Logger          logger.Logger
	StateRepository repository.UserStateRepository
}

type CancelScenarioUseCase struct {
	logger          logger.Logger
	stateRepository repository.UserStateRepository
}

func NewCancelScenarioUseCase(cfg *CancelScenarioUseCaseConfig) shared.UseCase[CancelScenarioUseCaseRequest, CancelScenarioUseCaseResponse] {
	return &CancelScenarioUseCase{
		logger:          cfg.Logger.With("layer", "usecase", "usecase", "CancelScenario"),
		stateRepository: cfg.StateRepository,
	}
}

func (uc *CancelScenarioUseCase) Execute(ctx context.Context, in CancelScenarioUseCaseRequest) (CancelScenarioUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID)

	stateOut, err := uc.stateRepository.Get(ctx, repository.UserStateRepositoryGetIn{UserID: in.UserID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get state", "error", err)
		return CancelScenarioUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to get state")
	}

	if !stateOut.Exists() || stateOut.UserState.SourceType == "" {
		return CancelScenarioUseCaseResponse{}, domainError.New(domainError.StatusConflict, "no active operations to cancel")
	}

	if _, err := uc.stateRepository.Delete(ctx, repository.UserStateRepositoryDeleteIn{UserID: in.UserID}); err != nil {
		logger.ErrorContext(ctx, "failed to delete state", "error", err)
		return CancelScenarioUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to delete state")
	}

	return CancelScenarioUseCaseResponse{}, nil
}
