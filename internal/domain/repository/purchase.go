package repository

import (
	"context"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/valueobject"
)

type PurchaseRepositorySaveIn struct {
	Purchase entity.Purchase
}

type PurchaseRepositorySaveOut struct{}

type PurchaseRepositoryGetByUserIDIn struct {
	UserID int64
	Limit  int
	Offset int
}

type PurchaseRepositoryGetByUserIDOut struct {
	Purchases []entity.Purchase
	Total     int
}

func (o *PurchaseRepositoryGetByUserIDOut) Exists() bool {
	return o.Purchases != nil
}

type PurchaseRepositoryGetByIDsIn struct {
	IDs []valueobject.UUID
}

type PurchaseRepositoryGetByIDsOut struct {
	Purchases []entity.Purchase
}

func (o *PurchaseRepositoryGetByIDsOut) Exists() bool {
	return o.Purchases != nil
}

type PurchaseRepositoryDeleteIn struct {
	UserID int64
	ID     valueobject.UUID
}

type PurchaseRepositoryDeleteOut struct{}

type PurchaseRepository interface {
	Save(ctx context.Context, in PurchaseRepositorySaveIn) (PurchaseRepositorySaveOut, error)
	GetByUserID(ctx context.Context, in PurchaseRepositoryGetByUserIDIn) (PurchaseRepositoryGetByUserIDOut, error)
	GetByIDs(ctx context.Context, in PurchaseRepositoryGetByIDsIn) (PurchaseRepositoryGetByIDsOut, error)
	Delete(ctx context.Context, in PurchaseRepositoryDeleteIn) (PurchaseRepositoryDeleteOut, error)
}
