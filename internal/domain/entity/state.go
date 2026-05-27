package entity

import "time"

type DraftItem struct {
	Name      string
	Quantity  float64
	UnitPrice float64
	Category  Category
}

type UserState struct {
	UserID int64

	DraftItems []DraftItem
	SourceType SourceType

	PurchaseDate time.Time
	Organisation string
}
