package category

import (
	"context"
	"fmt"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/repository"
	"fms-project/internal/infrastructure/logger"
	"fms-project/internal/infrastructure/storage/postgres"
	storageModel "fms-project/internal/infrastructure/storage/postgres/model"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CategoryRepositoryConfig struct {
	Logger logger.Logger
	Client *postgres.Client
}

type CategoryRepository struct {
	logger logger.Logger
	client *postgres.Client
}

func NewCategoryRepository(cfg *CategoryRepositoryConfig) repository.CategoryRepository {
	return &CategoryRepository{
		logger: cfg.Logger.With("layer", "repository", "repository", "Category"),
		client: cfg.Client,
	}
}

func (r *CategoryRepository) Create(ctx context.Context, in repository.CategoryRepositoryCreateIn) (repository.CategoryRepositoryCreateOut, error) {
	model := storageModel.CategoryFromEntity(in.Category)

	_, err := postgres.CheckTx(ctx, r.client).NewInsert().
		Model(&model).
		Exec(ctx)
	if err != nil {
		return repository.CategoryRepositoryCreateOut{}, err
	}

	return repository.CategoryRepositoryCreateOut{}, nil
}

func (r *CategoryRepository) GetByIDs(ctx context.Context, in repository.CategoryRepositoryGetByIDsIn) (repository.CategoryRepositoryGetByIDsOut, error) {
	ids := make([]uuid.UUID, 0, len(in.IDs))
	for _, id := range in.IDs {
		ids = append(ids, id.Value())
	}

	var models []storageModel.Category
	err := postgres.CheckTx(ctx, r.client).NewSelect().
		Model(&models).
		Where("id IN (?)", bun.List(ids)).
		Scan(ctx)
	if err != nil {
		return repository.CategoryRepositoryGetByIDsOut{}, err
	}

	if len(models) == 0 {
		return repository.CategoryRepositoryGetByIDsOut{Categories: nil}, nil
	}

	categories := make([]entity.Category, 0, len(models))
	for _, m := range models {
		category, err := storageModel.CategoryToEntity(m)
		if err != nil {
			return repository.CategoryRepositoryGetByIDsOut{}, fmt.Errorf("mapping category: %w", err)
		}
		categories = append(categories, category)
	}

	return repository.CategoryRepositoryGetByIDsOut{Categories: categories}, nil
}

func (r *CategoryRepository) List(ctx context.Context, _ repository.CategoryRepositoryListIn) (repository.CategoryRepositoryListOut, error) {
	var models []storageModel.Category
	err := postgres.CheckTx(ctx, r.client).NewSelect().
		Model(&models).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return repository.CategoryRepositoryListOut{}, err
	}

	if len(models) == 0 {
		return repository.CategoryRepositoryListOut{Categories: nil}, nil
	}

	categories := make([]entity.Category, 0, len(models))
	for _, m := range models {
		category, err := storageModel.CategoryToEntity(m)
		if err != nil {
			return repository.CategoryRepositoryListOut{}, fmt.Errorf("mapping category: %w", err)
		}
		categories = append(categories, category)
	}

	return repository.CategoryRepositoryListOut{Categories: categories}, nil
}
