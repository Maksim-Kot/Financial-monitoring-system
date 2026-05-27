package telegram

import (
	"context"
	"fmt"
	"sync"

	"fms-project/internal/application/container"
	"fms-project/internal/infrastructure/config"
	"fms-project/internal/infrastructure/logger"
	"fms-project/internal/presentation/telegram/handler"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotConfig struct {
	Config    *config.Config
	Logger    logger.Logger
	Container *container.Container
}

type Bot struct {
	bot      *tgbotapi.BotAPI
	logger   logger.Logger
	handler  *handler.Handler
	updates  tgbotapi.UpdatesChannel
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewBot(cfg *BotConfig) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Config.Services.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to init telegram bot: %w", err)
	}
	bot.Debug = false

	handler := handler.NewHandler(&handler.HandlerConfig{
		Logger:   cfg.Logger,
		Bot:      bot,
		UseCases: cfg.Container.Usecases,
	})

	return &Bot{
		bot:      bot,
		logger:   cfg.Logger.With("layer", "presentation", "component", "Bot"),
		handler:  handler,
		stopChan: make(chan struct{}),
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("bot authorized", "username", b.bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	b.updates = b.bot.GetUpdatesChan(updateConfig)

	b.wg.Add(1)
	go b.processUpdates(ctx)

	return nil
}

func (b *Bot) Stop() {
	b.logger.Info("stopping bot updates channel")
	b.bot.StopReceivingUpdates()

	close(b.stopChan)
	b.wg.Wait()

	b.logger.Info("bot stopped gracefully")
}

func (b *Bot) processUpdates(ctx context.Context) {
	defer b.wg.Done()

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("context cancelled. stopping update processing")
			return
		case <-b.stopChan:
			b.logger.Info("stop signal received. stopping update processing")
			return
		case update, ok := <-b.updates:
			if !ok {
				b.logger.Info("updates channel closed")
				return
			}
			b.wg.Add(1)
			go b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	defer b.wg.Done()

	b.handler.Handle(ctx, update)
}
