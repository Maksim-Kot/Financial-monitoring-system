package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Purchase struct {
	bun.BaseModel `bun:"table:purchases,alias:p"`

	ID               uuid.UUID `bun:"id,pk,type:uuid"`
	UserID           int64     `bun:"user_id,notnull"`
	PurchaseDate     time.Time `bun:"purchase_date,notnull"`
	OrganisationName string    `bun:"organisation_name,notnull"`
	Description      string    `bun:"description"`
	SourceType       string    `bun:"source_type,notnull"`
	CreatedAt        time.Time `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt        time.Time `bun:"updated_at,nullzero,notnull,default:now(),skipupdate"`
}

type Expense struct {
	bun.BaseModel `bun:"table:expenses,alias:e"`

	ID         uuid.UUID `bun:"id,pk,type:uuid"`
	PurchaseID uuid.UUID `bun:"purchase_id,notnull,type:uuid"`
	Name       string    `bun:"name,notnull"`
	Quantity   float64   `bun:"quantity,notnull"`
	UnitPrice  int64     `bun:"unit_price_minor,notnull"`
	TotalPrice int64     `bun:"total_price_minor,notnull"`
	Currency   string    `bun:"currency,notnull"`
	CategoryID uuid.UUID `bun:"category_id,notnull,type:uuid"`
	Category   *Category `bun:"rel:belongs-to,join:category_id=id"`
	CreatedAt  time.Time `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt  time.Time `bun:"updated_at,nullzero,notnull,default:now(),skipupdate"`
}

type Organisation struct {
	bun.BaseModel `bun:"table:organisations,alias:o"`

	ID        uuid.UUID `bun:"id,pk,type:uuid"`
	UserID    int64     `bun:"user_id,notnull"`
	Name      string    `bun:"name,notnull"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:now(),skipupdate"`
}

type Category struct {
	bun.BaseModel `bun:"table:categories,alias:c"`

	ID        uuid.UUID `bun:"id,pk,type:uuid"`
	Name      string    `bun:"name,notnull"`
	Icon      string    `bun:"icon"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:now()"`
}
