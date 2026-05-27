package container

import (
	"time"

	"fms-project/internal/domain/gateway"
	"fms-project/internal/infrastructure/config"
	impl "fms-project/internal/infrastructure/gateway"
	"fms-project/internal/infrastructure/logger"
)

type GatewaysContainer struct {
	PhotoParser        gateway.PhotoParserGateway
	TextParser         gateway.TextParserGateway
	CategoryClassifier gateway.CategoryClassifierGateway
}

type GatewaysContainerConfig struct {
	Config *config.Config
	Logger logger.Logger
}

func NewGatewaysContainer(cfg *GatewaysContainerConfig) *GatewaysContainer {
	groqService := impl.NewGroqService(impl.GroqServiceConfig{
		Logger:  cfg.Logger,
		BaseURL: cfg.Config.Services.GroqServiceURL,
		APIKey:  cfg.Config.Services.GroqServiceAPIKey,
		Timeout: 10 * time.Second,
	})

	return &GatewaysContainer{
		PhotoParser:        groqService,
		TextParser:         groqService,
		CategoryClassifier: groqService,
	}
}
