package handler

import (
	"context"

	"fms-project/internal/application/usecase"
	domainError "fms-project/internal/domain/shared/domain-error"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleCommand(ctx context.Context, userID, chatID int64, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		h.cmdStart(chatID)
	case "text":
		h.cmdText(chatID, userID)
	case "photo":
		h.cmdPhoto(chatID, userID)
	case "manual":
		h.cmdManual(chatID, userID)
	case "list":
		h.cmdList(ctx, userID, chatID)
	case "cancel":
		h.cmdCancel(ctx, userID, chatID)
	}
}

func (h *Handler) cmdStart(chatID int64) {
	h.sendMessage(chatID, "Привет! Я помогу вести учет покупок.\n\nДоступные команды:\n/text - добавить покупку текстом\n/photo - добавить покупку по фото чека\n/manual - добавить покупку вручную по позициям\n/list - посмотреть список покупок\n/cancel - отменить текущую операцию")
}

func (h *Handler) cmdText(chatID, userID int64) {
	h.state.SetStep(userID, stepWaitText)
	h.sendMessage(chatID, "Введите текст с описанием покупки.\nНапример: Молоко за 2.50 x2, Хлеб 1.20")
}

func (h *Handler) cmdPhoto(chatID, userID int64) {
	h.state.SetStep(userID, stepWaitPhoto)
	h.sendMessage(chatID, "Отправьте фото с чеком")
}

func (h *Handler) cmdManual(chatID, userID int64) {
	h.state.SetStep(userID, stepWaitManualItem)
	h.state.ClearManualItemsData(userID)
	h.sendMessage(chatID, "Введите позицию в формате:\nнаименование; количество; цена за единицу\n\nНапример: Молоко; 2; 1.50")
}

func (h *Handler) cmdList(ctx context.Context, userID, chatID int64) {
	out, err := h.useCases.ListScenario.Execute(ctx, usecase.ListScenarioUseCaseRequest{
		UserID: userID,
		Limit:  purchasesPerPage,
		Offset: 0,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to list scenario", "error", err)
		h.sendMessage(chatID, "Произошла ошибка. Попробуйте позже")
		return
	}

	message := buildPurchasesMessage(out.Purchases, 0, out.Total)
	var msg tgbotapi.MessageConfig
	if len(out.Purchases) != 0 {
		msg = tgbotapi.NewMessage(chatID, message)
		msg.ReplyMarkup = buildListKeyboard(0, out.Total)
	} else {
		msg = tgbotapi.NewMessage(chatID, message)
	}

	sent, err := h.bot.Send(msg)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to send list message", "error", err)
		return
	}

	h.state.SetList(userID, listPaginationState{
		MessageID: sent.MessageID,
		Offset:    0,
		Total:     out.Total,
	})
}

func (h *Handler) cmdCancel(ctx context.Context, userID, chatID int64) {
	_, err := h.useCases.CancelScenario.Execute(ctx, usecase.CancelScenarioUseCaseRequest{
		UserID: userID,
	})
	if err != nil {
		if domainError.IsDomain(err) {
			switch domainError.Extract(err).Status {
			case domainError.StatusConflict:
				if h.state.GetStep(userID) == stepIdle {
					h.sendMessageWithRemoveKeyboard(chatID, "Нет активных операций для отмены")
					return
				}
			default:
				h.logger.ErrorContext(ctx, "failed to cancel scenario", "error", err)
				h.sendMessage(chatID, "Произошла ошибка. Попробуйте позже")
				return
			}
		}
	}

	h.state.ClearStep(userID)
	h.sendMessageWithRemoveKeyboard(chatID, "Операция отменена")
}
