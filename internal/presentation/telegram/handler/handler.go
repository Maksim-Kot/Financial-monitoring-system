package handler

import (
	"context"
	"fms-project/internal/application/container"
	"fms-project/internal/infrastructure/adapter"
	"fms-project/internal/infrastructure/logger"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	purchasesPerPage      = 3
	maxSavedOrganisations = 10
	editPurchasesPerPage  = 1 // Show 1 purchase at a time for editing
	editJumpStep          = 5 // Jump 5 positions forward/backward
)

type HandlerConfig struct {
	Logger   logger.Logger
	Bot      *tgbotapi.BotAPI
	UseCases *container.UseCasesContainer
}

type Handler struct {
	logger     logger.Logger
	bot        *tgbotapi.BotAPI
	useCases   *container.UseCasesContainer
	httpClient *adapter.HttpClient

	state            *State
	callbackHandlers map[callbackAction]callbackHandler
}

func NewHandler(cfg *HandlerConfig) *Handler {
	h := &Handler{
		logger:   cfg.Logger.With("layer", "presentation", "component", "Handler"),
		bot:      cfg.Bot,
		useCases: cfg.UseCases,
		httpClient: adapter.NewHttpClient().
			WithTimeout(20 * time.Second).
			WithRetries(2),
		state: NewState(),
	}
	h.callbackHandlers = h.buildCallbackHandlers()
	return h
}

func (h *Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		h.handleCallbackQuery(ctx, update.CallbackQuery)
		return
	}

	if update.Message == nil || update.Message.From == nil {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	if update.Message.IsCommand() {
		h.handleCommand(ctx, userID, chatID, update.Message)
		return
	}

	h.handleStep(ctx, userID, chatID, update.Message)
}
