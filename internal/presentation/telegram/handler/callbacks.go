package handler

import (
	"context"

	"fms-project/internal/application/usecase"
	domainError "fms-project/internal/domain/shared/domain-error"
	"fms-project/internal/domain/valueobject"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type callbackAction string

const (
	callbackListPrev   callbackAction = "list_prev"
	callbackListNext   callbackAction = "list_next"
	callbackListFinish callbackAction = "list_finish"
	callbackSaveOrgYes callbackAction = "save_org_yes"
	callbackSaveOrgNo  callbackAction = "save_org_no"
	callbackAddMoreYes callbackAction = "add_more_yes"
	callbackAddMoreNo  callbackAction = "add_more_no"

	// Analytics callbacks
	callbackStatsDay         callbackAction = "stats_day"
	callbackStatsWeek        callbackAction = "stats_week"
	callbackStatsMonth       callbackAction = "stats_month"
	callbackStatsHalfYear    callbackAction = "stats_half_year"
	callbackStatsDetailedYes callbackAction = "stats_detailed_yes"
	callbackStatsBack        callbackAction = "stats_back"
	callbackStatsOtherPeriod callbackAction = "stats_other_period"
	callbackStatsClose       callbackAction = "stats_close"
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
	case callbackAddMoreYes, callbackAddMoreNo:
		h.handleAddMoreCallback(ctx, q, userID, chatID, messageID, action)
	case callbackStatsDay, callbackStatsWeek, callbackStatsMonth, callbackStatsHalfYear:
		h.handleStatsPeriodCallback(ctx, q, userID, chatID, messageID, action)
	case callbackStatsDetailedYes:
		h.handleStatsDetailedCallback(ctx, q, userID, chatID, messageID)
	case callbackStatsBack:
		h.handleStatsBackCallback(ctx, q, userID, chatID, messageID)
	case callbackStatsOtherPeriod:
		h.handleStatsOtherPeriodCallback(ctx, q, userID, chatID, messageID)
	case callbackStatsClose:
		h.handleStatsCloseCallback(ctx, q, userID, chatID, messageID)
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

func (h *Handler) handleAddMoreCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int, action callbackAction) {
	data, exists := h.state.GetManualItemsData(userID)
	if !exists || data.MessageID != messageID {
		h.answerCallback(q.ID, "Сессия устарела")
		return
	}

	if action == callbackAddMoreYes {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "Введите следующую позицию в формате:\nнаименование; количество; цена за единицу")
		edit.ReplyMarkup = nil
		h.bot.Send(edit)

		h.state.SetStep(userID, stepWaitManualItem)
	} else {
		h.state.ClearManualItemsData(userID)

		_, err := h.useCases.ClassifyCategory.Execute(ctx, usecase.ClassifyCategoryUseCaseRequest{
			UserID: userID,
		})
		if err != nil {
			h.logger.ErrorContext(ctx, "failed to classify categories", "error", err)
		}

		edit := tgbotapi.NewEditMessageText(chatID, messageID, "Позиции добавлены! Теперь введите дату покупки в формате ДД.ММ.ГГГГ")
		edit.ReplyMarkup = nil
		h.bot.Send(edit)

		h.state.SetStep(userID, stepWaitDate)
	}
}

// ==================== Analytics Callback Handlers ====================

func (h *Handler) handleStatsPeriodCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int, action callbackAction) {
	var periodType valueobject.PeriodType
	switch action {
	case callbackStatsDay:
		periodType = valueobject.PeriodTypeDay
	case callbackStatsWeek:
		periodType = valueobject.PeriodTypeWeek
	case callbackStatsMonth:
		periodType = valueobject.PeriodTypeMonth
	case callbackStatsHalfYear:
		periodType = valueobject.PeriodTypeHalfYear
	default:
		h.answerCallback(q.ID, "Неизвестный период")
		return
	}

	// Get Level 1 analytics (Summary)
	out, err := h.useCases.GetAnalytics.Execute(ctx, usecase.GetAnalyticsUseCaseRequest{
		UserID:     userID,
		PeriodType: periodType,
		Detailed:   false,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get analytics", "error", err)
		h.answerCallback(q.ID, "Ошибка получения данных")
		return
	}

	// Save state for later
	h.state.SetAnalytics(userID, analyticsState{
		MessageID:  messageID,
		PeriodType: periodType,
		Summary:    out.Summary,
	})

	// Display summary with prompt for detailed view
	message := buildSummaryMessage(out.Summary)
	keyboard := buildDetailedPromptKeyboard()

	edit := tgbotapi.NewEditMessageText(chatID, messageID, message)
	edit.ReplyMarkup = &keyboard
	h.bot.Send(edit)
	h.answerCallback(q.ID, "")
}

func (h *Handler) handleStatsDetailedCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int) {
	state, exists := h.state.GetAnalytics(userID)
	if !exists || state.MessageID != messageID {
		h.answerCallback(q.ID, "Сессия устарела")
		return
	}

	// Get Level 2 analytics (Detailed)
	out, err := h.useCases.GetAnalytics.Execute(ctx, usecase.GetAnalyticsUseCaseRequest{
		UserID:     userID,
		PeriodType: state.PeriodType,
		Detailed:   true,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get detailed analytics", "error", err)
		h.answerCallback(q.ID, "Ошибка получения данных")
		return
	}

	if out.Detailed == nil {
		// No detailed data available
		keyboard := buildNoDataKeyboard()
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "Нет данных для подробного анализа")
		edit.ReplyMarkup = &keyboard
		h.bot.Send(edit)
		h.answerCallback(q.ID, "")
		return
	}

	// Display detailed report
	message := buildDetailedMessage(*out.Detailed)
	keyboard := buildDetailedNavigationKeyboard()

	edit := tgbotapi.NewEditMessageText(chatID, messageID, message)
	edit.ReplyMarkup = &keyboard
	h.bot.Send(edit)
	h.answerCallback(q.ID, "")
}

func (h *Handler) handleStatsBackCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int) {
	_ = ctx
	state, exists := h.state.GetAnalytics(userID)
	if !exists || state.MessageID != messageID {
		h.answerCallback(q.ID, "Сессия устарела")
		return
	}

	// Go back to summary view
	message := buildSummaryMessage(state.Summary)
	keyboard := buildDetailedPromptKeyboard()

	edit := tgbotapi.NewEditMessageText(chatID, messageID, message)
	edit.ReplyMarkup = &keyboard
	h.bot.Send(edit)
	h.answerCallback(q.ID, "")
}

func (h *Handler) handleStatsOtherPeriodCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int) {
	_ = ctx
	// Clear current state and show period selection
	h.state.ClearAnalytics(userID)

	keyboard := buildPeriodSelectionKeyboard()
	edit := tgbotapi.NewEditMessageText(chatID, messageID, "📊 Выберите период для аналитики:")
	edit.ReplyMarkup = &keyboard
	h.bot.Send(edit)
	h.answerCallback(q.ID, "")
}

func (h *Handler) handleStatsCloseCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int) {
	_ = ctx
	h.state.ClearAnalytics(userID)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, "Аналитика закрыта")
	edit.ReplyMarkup = nil
	h.bot.Send(edit)
	h.answerCallback(q.ID, "")
}

func (h *Handler) answerCallback(callbackID, text string) {
	cfg := tgbotapi.NewCallback(callbackID, text)
	h.bot.Request(cfg)
}
