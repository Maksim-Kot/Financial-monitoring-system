package repository

import (
	"context"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/valueobject"
)

type OrganisationRepositoryCreateIn struct {
	Organisation entity.Organisation
}

type OrganisationRepositoryCreateOut struct{}

type OrganisationRepositoryListIn struct {
	UserID int64
}

type OrganisationRepositoryListOut struct {
	Organisations []entity.Organisation
}

func (o *OrganisationRepositoryListOut) Exists() bool {
	return o.Organisations != nil
}

type OrganisationRepositoryDeleteIn struct {
	UserID int64
	ID     valueobject.UUID
}

type OrganisationRepositoryDeleteOut struct{}

type OrganisationRepository interface {
	Create(ctx context.Context, in OrganisationRepositoryCreateIn) (OrganisationRepositoryCreateOut, error)
	List(ctx context.Context, in OrganisationRepositoryListIn) (OrganisationRepositoryListOut, error)
	Delete(ctx context.Context, in OrganisationRepositoryDeleteIn) (OrganisationRepositoryDeleteOut, error)
}
