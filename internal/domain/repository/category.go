package repository

import (
	"context"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/valueobject"
)

type CategoryRepositoryCreateIn struct {
	Category entity.Category
}

type CategoryRepositoryCreateOut struct{}

type CategoryRepositoryGetByIDsIn struct {
	IDs []valueobject.UUID
}

type CategoryRepositoryGetByIDsOut struct {
	Categories []entity.Category
}

func (o *CategoryRepositoryGetByIDsOut) Exists() bool {
	return o.Categories != nil
}

type CategoryRepositoryListIn struct{}

type CategoryRepositoryListOut struct {
	Categories []entity.Category
}

func (o *CategoryRepositoryListOut) Exists() bool {
	return o.Categories != nil
}

type CategoryRepository interface {
	Create(ctx context.Context, in CategoryRepositoryCreateIn) (CategoryRepositoryCreateOut, error)
	GetByIDs(ctx context.Context, in CategoryRepositoryGetByIDsIn) (CategoryRepositoryGetByIDsOut, error)
	List(ctx context.Context, in CategoryRepositoryListIn) (CategoryRepositoryListOut, error)
}
