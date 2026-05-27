package purchase

import (
	"context"
	"fmt"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/repository"
	"fms-project/internal/domain/valueobject"
	"fms-project/internal/infrastructure/logger"
	"fms-project/internal/infrastructure/storage/postgres"
	storageModel "fms-project/internal/infrastructure/storage/postgres/model"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PurchaseRepositoryConfig struct {
	Logger logger.Logger
	Client *postgres.Client
}

type PurchaseRepository struct {
	logger logger.Logger
	client *postgres.Client
}

func NewPurchaseRepository(cfg *PurchaseRepositoryConfig) repository.PurchaseRepository {
	return &PurchaseRepository{
		logger: cfg.Logger.With("layer", "repository", "repository", "Purchase"),
		client: cfg.Client,
	}
}

func (r *PurchaseRepository) Save(ctx context.Context, in repository.PurchaseRepositorySaveIn) (repository.PurchaseRepositorySaveOut, error) {
	db := postgres.CheckTx(ctx, r.client)

	exists, err := r.exists(ctx, db, in.Purchase.ID)
	if err != nil {
		return repository.PurchaseRepositorySaveOut{}, err
	}

	if !exists {
		if err := r.insertPurchase(ctx, db, &in.Purchase); err != nil {
			return repository.PurchaseRepositorySaveOut{}, err
		}

		if err := r.insertExpenses(ctx, db, &in.Purchase); err != nil {
			return repository.PurchaseRepositorySaveOut{}, err
		}

		return repository.PurchaseRepositorySaveOut{}, nil
	}

	if err := r.updatePurchase(ctx, db, &in.Purchase); err != nil {
		return repository.PurchaseRepositorySaveOut{}, err
	}

	if err := r.syncExpenses(ctx, db, &in.Purchase); err != nil {
		return repository.PurchaseRepositorySaveOut{}, err
	}

	return repository.PurchaseRepositorySaveOut{}, nil
}

func (r *PurchaseRepository) GetByUserID(ctx context.Context, in repository.PurchaseRepositoryGetByUserIDIn) (repository.PurchaseRepositoryGetByUserIDOut, error) {
	db := postgres.CheckTx(ctx, r.client)

	total, err := db.NewSelect().
		Model((*storageModel.Purchase)(nil)).
		Where("user_id = ?", in.UserID).
		Count(ctx)
	if err != nil {
		return repository.PurchaseRepositoryGetByUserIDOut{}, err
	}

	var models []storageModel.Purchase
	query := db.NewSelect().
		Model(&models).
		Where("user_id = ?", in.UserID).
		Order("purchase_date DESC", "created_at DESC")

	if in.Limit > 0 {
		query = query.Limit(in.Limit)
	}
	if in.Offset > 0 {
		query = query.Offset(in.Offset)
	}

	if err := query.Scan(ctx); err != nil {
		return repository.PurchaseRepositoryGetByUserIDOut{}, err
	}

	if len(models) == 0 {
		return repository.PurchaseRepositoryGetByUserIDOut{
			Purchases: nil,
			Total:     total,
		}, nil
	}

	purchases := make([]entity.Purchase, 0, len(models))
	for _, m := range models {
		purchase, err := storageModel.PurchaseToEntity(m)
		if err != nil {
			return repository.PurchaseRepositoryGetByUserIDOut{}, fmt.Errorf("mapping purchase: %w", err)
		}
		purchases = append(purchases, purchase)
	}

	return repository.PurchaseRepositoryGetByUserIDOut{
		Purchases: purchases,
		Total:     total,
	}, nil
}

func (r *PurchaseRepository) GetByIDs(ctx context.Context, in repository.PurchaseRepositoryGetByIDsIn) (repository.PurchaseRepositoryGetByIDsOut, error) {
	db := postgres.CheckTx(ctx, r.client)

	ids := make([]uuid.UUID, 0, len(in.IDs))
	for _, id := range in.IDs {
		ids = append(ids, id.Value())
	}

	var models []storageModel.Purchase
	err := db.NewSelect().
		Model(&models).
		Where("id IN (?)", bun.List(ids)).
		Scan(ctx)
	if err != nil {
		return repository.PurchaseRepositoryGetByIDsOut{}, err
	}

	if len(models) == 0 {
		return repository.PurchaseRepositoryGetByIDsOut{Purchases: nil}, nil
	}

	purchases := make([]entity.Purchase, 0, len(models))
	for _, m := range models {
		purchase, err := storageModel.PurchaseToEntity(m)
		if err != nil {
			return repository.PurchaseRepositoryGetByIDsOut{}, fmt.Errorf("mapping purchase: %w", err)
		}

		expenses, err := r.getExpenses(ctx, db, purchase.ID)
		if err != nil {
			return repository.PurchaseRepositoryGetByIDsOut{}, err
		}

		for _, e := range expenses {
			purchase.AddExpense(e)
		}
		purchases = append(purchases, purchase)
	}

	return repository.PurchaseRepositoryGetByIDsOut{Purchases: purchases}, nil
}

func (r *PurchaseRepository) Delete(ctx context.Context, in repository.PurchaseRepositoryDeleteIn) (repository.PurchaseRepositoryDeleteOut, error) {
	db := postgres.CheckTx(ctx, r.client)

	_, err := db.NewDelete().
		Model((*storageModel.Purchase)(nil)).
		Where("id = ?", in.ID.Value()).
		Where("user_id = ?", in.UserID).
		Exec(ctx)
	if err != nil {
		return repository.PurchaseRepositoryDeleteOut{}, err
	}

	return repository.PurchaseRepositoryDeleteOut{}, nil
}

func (r *PurchaseRepository) exists(ctx context.Context, db bun.IDB, id valueobject.UUID) (bool, error) {
	count, err := db.NewSelect().
		Model((*storageModel.Purchase)(nil)).
		Where("id = ?", id.Value()).
		Count(ctx)

	return count > 0, err
}

func (r *PurchaseRepository) insertPurchase(ctx context.Context, db bun.IDB, p *entity.Purchase) error {
	model := storageModel.PurchaseFromEntity(*p)

	_, err := db.NewInsert().
		Model(&model).
		Exec(ctx)

	return err
}

func (r *PurchaseRepository) updatePurchase(ctx context.Context, db bun.IDB, p *entity.Purchase) error {
	model := storageModel.PurchaseFromEntity(*p)

	_, err := db.NewUpdate().
		Model(&model).
		Where("id = ?", p.ID.Value()).
		Column("purchase_date", "organisation_name", "description", "source_type").
		Exec(ctx)

	return err
}

func (r *PurchaseRepository) insertExpenses(ctx context.Context, db bun.IDB, p *entity.Purchase) error {
	expenses := p.Expenses()

	if len(expenses) == 0 {
		return nil
	}

	models := make([]storageModel.Expense, 0, len(expenses))

	for _, e := range expenses {
		models = append(models, storageModel.ExpenseFromEntity(e, p.ID))
	}

	_, err := db.NewInsert().
		Model(&models).
		Exec(ctx)

	return err
}

func (r *PurchaseRepository) syncExpenses(ctx context.Context, db bun.IDB, p *entity.Purchase) error {
	var models []storageModel.Expense

	err := db.NewSelect().
		Model(&models).
		Where("purchase_id = ?", p.ID.Value()).
		Scan(ctx)
	if err != nil {
		return err
	}

	// map для быстрого доступа
	modelsMap := make(map[string]storageModel.Expense)
	for _, m := range models {
		modelsMap[m.ID.String()] = m
	}

	inputExpenses := p.Expenses()

	inputMap := make(map[string]entity.Expense)
	for _, e := range inputExpenses {
		inputMap[e.ID.String()] = e
	}

	for _, e := range inputExpenses {
		model := storageModel.ExpenseFromEntity(e, p.ID)

		if _, ok := modelsMap[e.ID.String()]; ok {
			_, err := db.NewUpdate().
				Model(&model).
				Where("id = ?", e.ID.String()).
				Exec(ctx)
			if err != nil {
				return err
			}

			continue
		}

		_, err := db.NewInsert().
			Model(&model).
			Exec(ctx)
		if err != nil {
			return err
		}
	}

	for _, model := range models {
		if _, ok := inputMap[model.ID.String()]; !ok {
			_, err := db.NewDelete().
				Model((*storageModel.Expense)(nil)).
				Where("id = ?", model.ID.String()).
				Exec(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *PurchaseRepository) getExpenses(ctx context.Context, db bun.IDB, purchaseID valueobject.UUID) ([]entity.Expense, error) {
	var rows []storageModel.Expense

	err := db.NewSelect().
		Model(&rows).
		Relation("Category").
		Where("e.purchase_id = ?", purchaseID.Value()).
		Order("e.created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	expenses := make([]entity.Expense, 0, len(rows))

	for _, row := range rows {
		if row.Category == nil {
			return nil, fmt.Errorf("category not found for expense %s", row.ID)
		}

		e, err := storageModel.ExpenseToEntity(row, *row.Category)
		if err != nil {
			return nil, fmt.Errorf("mapping expense: %w", err)
		}

		expenses = append(expenses, e)
	}

	return expenses, nil
}
