package entity

import (
	"fms-project/internal/domain/valueobject"
)

type Expense struct {
	ID        valueobject.UUID
	Name      string
	Quantity  float64
	UnitPrice valueobject.MoneyAmount
	Category  Category
}

func NewExpense(purchaseID valueobject.UUID) Expense {
	return Expense{
		ID: valueobject.NewRandom(),
	}
}

func (e *Expense) SetName(name string) {
	e.Name = name
}

func (e *Expense) SetQuantity(quantity float64) {
	e.Quantity = quantity
}

func (e *Expense) SetUnitPrice(unitPrice valueobject.MoneyAmount) {
	e.UnitPrice = unitPrice
}

func (e *Expense) SetCategory(category Category) {
	e.Category = category
}

func (e Expense) TotalPrice() (valueobject.MoneyAmount, error) {
	return e.UnitPrice.MulFloat64(e.Quantity)
}
