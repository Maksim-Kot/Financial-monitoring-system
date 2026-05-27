package entity

import (
	"fmt"
	"time"

	"fms-project/internal/domain/valueobject"
)

type SourceType string

const (
	SourceTypeManual SourceType = "manual"
	SourceTypeText   SourceType = "text"
	SourceTypePhoto  SourceType = "photo"
)

type Purchase struct {
	ID               valueobject.UUID
	UserID           int64
	PurchaseDate     time.Time
	OrganisationName string
	Description      string
	SourceType       SourceType

	expenses []Expense
}

func NewPurchase(userID int64, sourceType SourceType) Purchase {
	return Purchase{
		ID:         valueobject.NewRandom(),
		UserID:     userID,
		SourceType: sourceType,
		expenses:   make([]Expense, 0),
	}
}

func (p *Purchase) SetOrganisation(organisationName string) {
	p.OrganisationName = organisationName
}

func (p *Purchase) SetPurchaseDate(purchaseDate time.Time) {
	p.PurchaseDate = purchaseDate
}

func (p *Purchase) SetDescription(description string) {
	p.Description = description
}

func (p *Purchase) AddExpense(expense Expense) {
	p.expenses = append(p.expenses, expense)
}

func (p *Purchase) Expenses() []Expense {
	// return copy to protect invariants
	result := make([]Expense, len(p.expenses))
	copy(result, p.expenses)
	return result
}

func (p *Purchase) TotalPrice() (valueobject.MoneyAmount, error) {
	total, err := valueobject.NewZeroMoneyAmount(2, valueobject.MoneyAmountCurrencyBYN, nil)
	if err != nil {
		return valueobject.MoneyAmount{}, err
	}
	for _, expense := range p.expenses {
		expenseTotal, err := expense.TotalPrice()
		if err != nil {
			return valueobject.MoneyAmount{}, fmt.Errorf("failed to get expense total price: %w", err)
		}
		total, err = total.Add(expenseTotal)
		if err != nil {
			return valueobject.MoneyAmount{}, fmt.Errorf("failed to add expense total price: %w", err)
		}
	}
	return total, nil
}
