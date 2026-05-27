package gateway

import (
	"context"

	"fms-project/internal/domain/entity"
)

type PhotoParserGatewayIn struct {
	Photo []byte
}

type PhotoParserGatewayOut struct {
	Expenses []entity.DraftItem
}

type PhotoParserGateway interface {
	ParsePhoto(ctx context.Context, in PhotoParserGatewayIn) (PhotoParserGatewayOut, error)
}
