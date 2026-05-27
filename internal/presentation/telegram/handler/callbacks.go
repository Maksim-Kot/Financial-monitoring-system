package handler

import (
	"context"

	"fms-project/internal/application/usecase"
	domainError "fms-project/internal/domain/shared/domain-error"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type callbackAction string

const (
	callbackListPrev   callbackAction = "list_prev"
	callbackListNext   callbackAction = "list_next"
	callbackListFinish callbackAction = "list_finish"
	callbackSaveOrgYes callbackAction = "save_org_yes"
	callbackSaveOrgNo  callbackAction = "save_org_no"
)

func (h *Handler) handleCallbackQuery(ctx context.Context, q *tgbotapi.CallbackQuery) {
	if q.Message == nil {
		return
	}

	userID := q.From.ID
	chatID := q.Message.Chat.ID
	messageID := q.Message.MessageID

	action := callbackAction(q.Data)

	switch action {
	case callbackListPrev, callbackListNext, callbackListFinish:
		h.handleListCallback(ctx, q, userID, chatID, messageID, action)
	case callbackSaveOrgYes, callbackSaveOrgNo:
		h.handleSaveOrgCallback(ctx, q, userID, chatID, messageID, action)
	default:
		h.answerCallback(q.ID, "Неизвестная команда")
	}
}

func (h *Handler) handleListCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int, action callbackAction) {
	state, exists := h.state.GetList(userID)
	if !exists || state.MessageID != messageID {
		h.answerCallback(q.ID, "Сессия просмотра устарела")
		return
	}

	newOffset := state.Offset

	switch action {
	case callbackListPrev:
		newOffset = max(0, state.Offset-purchasesPerPage)
	case callbackListNext:
		newOffset = state.Offset + purchasesPerPage
	case callbackListFinish:
		h.state.ClearList(userID)
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "Просмотр списка покупок завершен")
		edit.ReplyMarkup = nil
		h.bot.Send(edit)
		h.answerCallback(q.ID, "Просмотр завершен")
		return
	}

	out, err := h.useCases.ListScenario.Execute(ctx, usecase.ListScenarioUseCaseRequest{
		UserID: userID,
		Limit:  purchasesPerPage,
		Offset: newOffset,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to list scenario", "error", err)
		h.answerCallback(q.ID, "Произошла ошибка")
		return
	}

	message := buildPurchasesMessage(out.Purchases, newOffset, out.Total)
	keyboard := buildListKeyboard(newOffset, out.Total)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, message)
	edit.ReplyMarkup = &keyboard
	edit.ParseMode = ""

	_, err = h.bot.Send(edit)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to edit message", "error", err)
		h.answerCallback(q.ID, "Ошибка обновления")
		return
	}

	h.state.SetList(userID, listPaginationState{
		MessageID: messageID,
		Offset:    newOffset,
		Total:     out.Total,
	})

	h.answerCallback(q.ID, "")
}

func (h *Handler) handleSaveOrgCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int, action callbackAction) {
	data, exists := h.state.GetSaveOrgData(userID)
	if !exists || data.MessageID != messageID {
		h.answerCallback(q.ID, "Сессия устарела")
		return
	}

	if action == callbackSaveOrgYes {
		_, err := h.useCases.SaveOrganisation.Execute(ctx, usecase.SaveOrganisationUseCaseRequest{
			UserID: userID,
			Name:   data.Name,
		})
		if err != nil {
			if domainError.HasStatus(err, domainError.StatusConflict) {
				edit := tgbotapi.NewEditMessageText(chatID, messageID, "Организация \""+data.Name+"\" уже есть в вашем списке.")
				edit.ReplyMarkup = nil
				h.bot.Send(edit)
				h.answerCallback(q.ID, "Уже сохранена")
			} else {
				h.logger.ErrorContext(ctx, "failed to save organisation", "error", err)
				h.answerCallback(q.ID, "Ошибка сохранения")
				return
			}
		} else {
			edit := tgbotapi.NewEditMessageText(chatID, messageID, "Организация \""+data.Name+"\" сохранена для быстрого выбора в будущем.")
			edit.ReplyMarkup = nil
			h.bot.Send(edit)
			h.answerCallback(q.ID, "Сохранено")
		}
	} else {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "Организация не сохранена.")
		edit.ReplyMarkup = nil
		h.bot.Send(edit)
		h.answerCallback(q.ID, "Пропущено")
	}

	h.state.ClearSaveOrgData(userID)
	h.finalizePurchase(ctx, userID, chatID)
}

func (h *Handler) answerCallback(callbackID, text string) {
	cfg := tgbotapi.NewCallback(callbackID, text)
	h.bot.Request(cfg)
}
