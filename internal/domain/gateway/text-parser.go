package gateway

import (
	"context"

	"fms-project/internal/domain/entity"
)

type TextParserGatewayIn struct {
	Text string
}

type TextParserGatewayOut struct {
	Expenses []entity.DraftItem
}

type TextParserGateway interface {
	ParseText(ctx context.Context, in TextParserGatewayIn) (TextParserGatewayOut, error)
}
