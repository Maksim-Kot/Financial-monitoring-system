package usecase

import (
	"context"
	"strings"

	"fms-project/internal/application/shared"
	"fms-project/internal/domain/repository"
	domainError "fms-project/internal/domain/shared/domain-error"
	"fms-project/internal/infrastructure/logger"
)

type AddOrganisationUseCaseRequest struct {
	UserID int64
	Name   string
}

type AddOrganisationUseCaseResponse struct {
}

type AddOrganisationUseCaseConfig struct {
	Logger                 logger.Logger
	StateRepository        repository.UserStateRepository
	OrganisationRepository repository.OrganisationRepository
}

type AddOrganisationUseCase struct {
	logger                 logger.Logger
	stateRepository        repository.UserStateRepository
	organisationRepository repository.OrganisationRepository
}

func NewAddOrganisationUseCase(cfg *AddOrganisationUseCaseConfig) shared.UseCase[AddOrganisationUseCaseRequest, AddOrganisationUseCaseResponse] {
	return &AddOrganisationUseCase{
		logger:                 cfg.Logger.With("layer", "usecase", "usecase", "AddOrganisation"),
		stateRepository:        cfg.StateRepository,
		organisationRepository: cfg.OrganisationRepository,
	}
}

func (uc *AddOrganisationUseCase) Execute(ctx context.Context, in AddOrganisationUseCaseRequest) (AddOrganisationUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID)

	organisationName := strings.TrimSpace(in.Name)
	if organisationName == "" {
		return AddOrganisationUseCaseResponse{}, domainError.New(domainError.StatusValidation, "organisation name is empty")
	}

	stateOut, err := uc.stateRepository.Get(ctx, repository.UserStateRepositoryGetIn{UserID: in.UserID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get state", "error", err)
		return AddOrganisationUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to get state")
	}
	if !stateOut.Exists() {
		return AddOrganisationUseCaseResponse{}, domainError.New(domainError.StatusNotFound, "state not found")
	}

	state := stateOut.UserState
	if state.UserID != in.UserID {
		return AddOrganisationUseCaseResponse{}, domainError.New(domainError.StatusConflict, "user ID mismatch")
	}

	if state.SourceType == "" {
		return AddOrganisationUseCaseResponse{}, domainError.New(domainError.StatusConflict, "source type is empty")
	}

	state.Organisation = organisationName
	if _, err := uc.stateRepository.Save(ctx, repository.UserStateRepositorySaveIn{UserState: state}); err != nil {
		logger.ErrorContext(ctx, "failed to save state", "error", err)
		return AddOrganisationUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to save state")
	}

	return AddOrganisationUseCaseResponse{}, nil
}
