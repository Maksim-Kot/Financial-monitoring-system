package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"fms-project/internal/application/bootstrap"
	"fms-project/internal/application/container"
	"fms-project/internal/infrastructure/config"
	"fms-project/internal/infrastructure/logger"
	"fms-project/internal/infrastructure/storage/postgres"
	"fms-project/internal/presentation/telegram"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	consoleLogger := logger.NewStdoutJSON("debug")

	cfg, err := config.Load()
	if err != nil {
		consoleLogger.Error("failed to load config", "error", err)
		return
	}

	logger := logger.NewStdoutJSON(cfg.Infrastructure.LogLevel)

	postgresClient, err := postgres.NewClient(postgres.ClientParams{
		DSN: cfg.Infrastructure.PostgresUri,
	})
	if err != nil {
		logger.Error("failed to init postgres client", "error", err)
		return
	}
	defer func() {
		if err := postgresClient.Close(); err != nil {
			logger.Error("failed to close postgres connection", "err", err)
		}
	}()

	if err := postgresClient.Migrate(ctx); err != nil {
		logger.Error("failed to migrate postgres", "error", err)
		return
	}

	container := container.NewContainer(ctx, &container.ContainerConfig{
		Config:   cfg,
		Logger:   logger,
		Postgres: postgresClient,
	})

	bootstrap := bootstrap.New(&bootstrap.BootstrapConfig{
		EnsureCategories: container.Usecases.EnsureCategories,
		Logger:           logger,
	})

	if err := bootstrap.Run(ctx); err != nil {
		logger.Error("failed to bootstrap", "error", err)
		return
	}

	bot, err := telegram.NewBot(&telegram.BotConfig{
		Config:    cfg,
		Logger:    logger,
		Container: container,
	})
	if err != nil {
		logger.Error("failed to init telegram bot", "error", err)
		return
	}

	if err := bot.Start(ctx); err != nil {
		logger.Error("failed to start bot", "error", err)
		return
	}

	logger.Info("bot successfully started")

	<-ctx.Done()
	logger.Info("shutdown signal received. initiating graceful shutdown")

	bot.Stop()

	logger.Info("graceful shutdown completed")
}
