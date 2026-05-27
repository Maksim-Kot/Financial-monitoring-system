package gateway

import (
	"context"

	"fms-project/internal/domain/entity"
)

type CategoryClassifierGatewayIn struct {
	Items      []entity.DraftItem
	Categories []entity.Category
}

type CategoryClassifierGatewayOut struct {
	Items []entity.DraftItem
}

type CategoryClassifierGateway interface {
	ClassifyCategories(ctx context.Context, in CategoryClassifierGatewayIn) (CategoryClassifierGatewayOut, error)
}
