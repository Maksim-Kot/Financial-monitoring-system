package usecase

import (
	"context"
	"strings"

	"fms-project/internal/application/shared"
	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/repository"
	domainError "fms-project/internal/domain/shared/domain-error"
	"fms-project/internal/domain/transaction"
	"fms-project/internal/infrastructure/logger"
)

type SaveOrganisationUseCaseRequest struct {
	UserID int64
	Name   string
}

type SaveOrganisationUseCaseResponse struct {
	Organisation entity.Organisation
}

type SaveOrganisationUseCaseConfig struct {
	Logger                 logger.Logger
	OrganisationRepository repository.OrganisationRepository
	TxManager              transaction.TxManager
}

type SaveOrganisationUseCase struct {
	logger                 logger.Logger
	organisationRepository repository.OrganisationRepository
	txManager              transaction.TxManager
}

func NewSaveOrganisationUseCase(cfg *SaveOrganisationUseCaseConfig) shared.UseCase[SaveOrganisationUseCaseRequest, SaveOrganisationUseCaseResponse] {
	return &SaveOrganisationUseCase{
		logger:                 cfg.Logger.With("layer", "usecase", "usecase", "SaveOrganisation"),
		organisationRepository: cfg.OrganisationRepository,
		txManager:              cfg.TxManager,
	}
}

func (uc *SaveOrganisationUseCase) Execute(ctx context.Context, in SaveOrganisationUseCaseRequest) (SaveOrganisationUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID)

	organisationName := strings.TrimSpace(in.Name)
	if organisationName == "" {
		return SaveOrganisationUseCaseResponse{}, domainError.New(domainError.StatusValidation, "organisation name is empty")
	}

	organisation := entity.NewOrganisation(in.UserID, organisationName)

	err := uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		listOut, err := uc.organisationRepository.List(ctx, repository.OrganisationRepositoryListIn{
			UserID: in.UserID,
		})
		if err != nil {
			logger.ErrorContext(ctx, "failed to list organisations", "error", err)
			return domainError.New(domainError.StatusInternal, "failed to list organisations")
		}

		searchName := strings.ToLower(organisationName)
		for _, org := range listOut.Organisations {
			if strings.ToLower(strings.TrimSpace(org.Name)) == searchName {
				return domainError.New(domainError.StatusConflict, "organisation already exists")
			}
		}

		_, err = uc.organisationRepository.Create(ctx, repository.OrganisationRepositoryCreateIn{
			Organisation: organisation,
		})
		if err != nil {
			logger.ErrorContext(ctx, "failed to create organisation", "error", err)
			return domainError.New(domainError.StatusInternal, "failed to create organisation")
		}

		return nil
	})

	return SaveOrganisationUseCaseResponse{Organisation: organisation}, err
}
