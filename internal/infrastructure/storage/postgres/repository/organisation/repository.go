package organisation

import (
	"context"
	"fmt"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/repository"
	"fms-project/internal/infrastructure/logger"
	"fms-project/internal/infrastructure/storage/postgres"
	storageModel "fms-project/internal/infrastructure/storage/postgres/model"
)

type OrganisationRepositoryConfig struct {
	Logger logger.Logger
	Client *postgres.Client
}

type OrganisationRepository struct {
	logger logger.Logger
	client *postgres.Client
}

func NewOrganisationRepository(cfg *OrganisationRepositoryConfig) repository.OrganisationRepository {
	return &OrganisationRepository{
		logger: cfg.Logger.With("layer", "repository", "repository", "Organisation"),
		client: cfg.Client,
	}
}

func (r *OrganisationRepository) Create(ctx context.Context, in repository.OrganisationRepositoryCreateIn) (repository.OrganisationRepositoryCreateOut, error) {
	model := storageModel.OrganisationFromEntity(in.Organisation)

	_, err := postgres.CheckTx(ctx, r.client).NewInsert().
		Model(&model).
		Exec(ctx)
	if err != nil {
		return repository.OrganisationRepositoryCreateOut{}, err
	}

	return repository.OrganisationRepositoryCreateOut{}, nil
}

func (r *OrganisationRepository) List(ctx context.Context, in repository.OrganisationRepositoryListIn) (repository.OrganisationRepositoryListOut, error) {
	var models []storageModel.Organisation
	err := postgres.CheckTx(ctx, r.client).NewSelect().
		Model(&models).
		Where("user_id = ?", in.UserID).
		Order("updated_at DESC").
		Scan(ctx)
	if err != nil {
		return repository.OrganisationRepositoryListOut{}, err
	}

	if len(models) == 0 {
		return repository.OrganisationRepositoryListOut{Organisations: nil}, nil
	}

	organisations := make([]entity.Organisation, 0, len(models))
	for _, m := range models {
		organisation, err := storageModel.OrganisationToEntity(m)
		if err != nil {
			return repository.OrganisationRepositoryListOut{}, fmt.Errorf("mapping organisation: %w", err)
		}
		organisations = append(organisations, organisation)
	}

	return repository.OrganisationRepositoryListOut{Organisations: organisations}, nil
}

func (r *OrganisationRepository) Delete(ctx context.Context, in repository.OrganisationRepositoryDeleteIn) (repository.OrganisationRepositoryDeleteOut, error) {
	_, err := postgres.CheckTx(ctx, r.client).NewDelete().
		Model((*storageModel.Organisation)(nil)).
		Where("id = ?", in.ID.Value()).
		Where("user_id = ?", in.UserID).
		Exec(ctx)
	if err != nil {
		return repository.OrganisationRepositoryDeleteOut{}, err
	}

	return repository.OrganisationRepositoryDeleteOut{}, nil
}
