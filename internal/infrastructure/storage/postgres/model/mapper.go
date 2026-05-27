package model

import (
	"fmt"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/valueobject"
)

const moneyDecimals uint8 = 2

// PurchaseFromEntity maps domain purchase to storage model
func PurchaseFromEntity(p entity.Purchase) Purchase {
	return Purchase{
		ID:               p.ID.Value(),
		UserID:           p.UserID,
		PurchaseDate:     p.PurchaseDate,
		OrganisationName: p.OrganisationName,
		Description:      p.Description,
		SourceType:       string(p.SourceType),
	}
}

// PurchaseToEntity maps storage model to domain purchase
func PurchaseToEntity(m Purchase) (entity.Purchase, error) {
	id, err := valueobject.NewUUID(m.ID.String())
	if err != nil {
		return entity.Purchase{}, fmt.Errorf("mapping purchase id: %w", err)
	}

	return entity.Purchase{
		ID:               id,
		UserID:           m.UserID,
		PurchaseDate:     m.PurchaseDate,
		OrganisationName: m.OrganisationName,
		Description:      m.Description,
		SourceType:       entity.SourceType(m.SourceType),
	}, nil
}

// ExpenseFromEntity maps domain expense to storage model
func ExpenseFromEntity(e entity.Expense, purchaseID valueobject.UUID) Expense {
	totalPrice, err := e.TotalPrice()
	if err != nil {
		totalPrice = e.UnitPrice
	}

	return Expense{
		ID:         e.ID.Value(),
		PurchaseID: purchaseID.Value(),
		Name:       e.Name,
		Quantity:   e.Quantity,
		UnitPrice:  e.UnitPrice.Int64(),
		TotalPrice: totalPrice.Int64(),
		Currency:   string(e.UnitPrice.Currency()),
		CategoryID: e.Category.ID.Value(),
	}
}

// ExpenseToEntity maps storage model to domain expense with category
func ExpenseToEntity(m Expense, category Category) (entity.Expense, error) {
	id, err := valueobject.NewUUID(m.ID.String())
	if err != nil {
		return entity.Expense{}, fmt.Errorf("mapping expense id: %w", err)
	}

	unitPrice, err := valueobject.NewMoneyAmountFromInt64(
		m.UnitPrice,
		moneyDecimals,
		valueobject.MoneyAmountCurrency(m.Currency),
		nil,
	)
	if err != nil {
		return entity.Expense{}, fmt.Errorf("mapping expense unit price: %w", err)
	}

	catEntity, err := CategoryToEntity(category)
	if err != nil {
		return entity.Expense{}, fmt.Errorf("mapping expense category: %w", err)
	}

	return entity.Expense{
		ID:        id,
		Name:      m.Name,
		Quantity:  m.Quantity,
		UnitPrice: unitPrice,
		Category:  catEntity,
	}, nil
}

// OrganisationFromEntity maps domain organisation to storage model
func OrganisationFromEntity(o entity.Organisation) Organisation {
	return Organisation{
		ID:     o.ID.Value(),
		UserID: o.UserID,
		Name:   o.Name,
	}
}

// OrganisationToEntity maps storage model to domain organisation
func OrganisationToEntity(m Organisation) (entity.Organisation, error) {
	id, err := valueobject.NewUUID(m.ID.String())
	if err != nil {
		return entity.Organisation{}, fmt.Errorf("mapping organisation id: %w", err)
	}

	return entity.Organisation{
		ID:     id,
		UserID: m.UserID,
		Name:   m.Name,
	}, nil
}

// CategoryFromEntity maps domain category to storage model
func CategoryFromEntity(c entity.Category) Category {
	return Category{
		ID:   c.ID.Value(),
		Name: c.Name,
		Icon: c.Icon,
	}
}

// CategoryToEntity maps storage model to domain category
func CategoryToEntity(m Category) (entity.Category, error) {
	id, err := valueobject.NewUUID(m.ID.String())
	if err != nil {
		return entity.Category{}, fmt.Errorf("mapping category id: %w", err)
	}

	return entity.Category{
		ID:   id,
		Name: m.Name,
		Icon: m.Icon,
	}, nil
}
