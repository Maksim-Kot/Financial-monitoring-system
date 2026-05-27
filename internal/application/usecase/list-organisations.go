package usecase

import (
	"context"

	"fms-project/internal/application/shared"
	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/repository"
	domainError "fms-project/internal/domain/shared/domain-error"
	"fms-project/internal/infrastructure/logger"
)

type ListOrganisationsUseCaseRequest struct {
	UserID int64
}

type ListOrganisationsUseCaseResponse struct {
	Organisations []entity.Organisation
}

type ListOrganisationsUseCaseConfig struct {
	Logger                 logger.Logger
	OrganisationRepository repository.OrganisationRepository
}

type ListOrganisationsUseCase struct {
	logger                 logger.Logger
	organisationRepository repository.OrganisationRepository
}

func NewListOrganisationsUseCase(cfg *ListOrganisationsUseCaseConfig) shared.UseCase[ListOrganisationsUseCaseRequest, ListOrganisationsUseCaseResponse] {
	return &ListOrganisationsUseCase{
		logger:                 cfg.Logger.With("layer", "usecase", "usecase", "ListUserOrganisations"),
		organisationRepository: cfg.OrganisationRepository,
	}
}

func (uc *ListOrganisationsUseCase) Execute(ctx context.Context, in ListOrganisationsUseCaseRequest) (ListOrganisationsUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID)

	listOut, err := uc.organisationRepository.List(ctx, repository.OrganisationRepositoryListIn{
		UserID: in.UserID,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to list organisations", "error", err)
		return ListOrganisationsUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to list organisations")
	}

	return ListOrganisationsUseCaseResponse{Organisations: listOut.Organisations}, nil
}
