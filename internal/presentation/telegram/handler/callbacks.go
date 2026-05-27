package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"fms-project/internal/application/usecase"
	"fms-project/internal/domain/entity"
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

	// Edit scenario callbacks (prefixes for dynamic data)
	callbackEditYearPrefix     = "edit_y_" // +year
	callbackEditMonthPrefix    = "edit_m_" // +year_month
	callbackEditBackToYears    = "edit_back_y"
	callbackEditPrev           = "edit_prev"
	callbackEditNext           = "edit_next"
	callbackEditJumpPrev       = "edit_jump_p"
	callbackEditJumpNext       = "edit_jump_n"
	callbackEditSelectPurchase = "edit_sel_p_" // +purchaseID
	callbackEditSelectExpense  = "edit_sel_e_" // +expenseID
	callbackEditCancelExpense  = "edit_can_exp"
	callbackEditFinish         = "edit_finish"
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
		// Check for dynamic edit callbacks
		if strings.HasPrefix(q.Data, callbackEditYearPrefix) {
			h.handleEditYearCallback(ctx, q, userID, chatID, messageID)
		} else if strings.HasPrefix(q.Data, callbackEditMonthPrefix) {
			h.handleEditMonthCallback(ctx, q, userID, chatID, messageID)
		} else if strings.HasPrefix(q.Data, callbackEditSelectPurchase) {
			h.handleEditSelectPurchaseCallback(ctx, q, userID, chatID, messageID)
		} else if strings.HasPrefix(q.Data, callbackEditSelectExpense) {
			h.handleEditSelectExpenseCallback(ctx, q, userID, chatID, messageID)
		} else if action == callbackEditBackToYears || action == callbackEditPrev ||
			action == callbackEditNext || action == callbackEditJumpPrev ||
			action == callbackEditJumpNext || action == callbackEditCancelExpense ||
			action == callbackEditFinish {
			h.handleEditNavigationCallback(ctx, q, userID, chatID, messageID, action)
		} else {
			h.answerCallback(q.ID, "Неизвестная команда")
		}
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

// ==================== Edit Scenario Callback Handlers ====================

func (h *Handler) handleEditYearCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int) {
	// Parse year from callback data
	prefix := callbackEditYearPrefix
	yearStr := strings.TrimPrefix(q.Data, prefix)
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		h.answerCallback(q.ID, "Некорректный год")
		return
	}

	// Get state and find months for this year
	state, exists := h.state.GetEditState(userID)
	if !exists {
		h.answerCallback(q.ID, "Сессия устарела")
		return
	}

	// Find months for selected year
	var months []int
	for _, period := range state.AvailablePeriods {
		if period.Year == year {
			months = period.Months
			break
		}
	}

	if len(months) == 0 {
		h.answerCallback(q.ID, "Нет данных для этого года")
		return
	}

	// Update state
	state.SelectedYear = year
	h.state.SetEditState(userID, state)
	h.state.SetStep(userID, stepEditSelectMonth)

	// Show months for selected year
	keyboard := buildMonthSelectionKeyboard(year, months)
	edit := tgbotapi.NewEditMessageText(chatID, messageID, fmt.Sprintf("📅 %d год. Выберите месяц:", year))
	edit.ReplyMarkup = &keyboard
	h.bot.Send(edit)
	h.answerCallback(q.ID, "")
}

func (h *Handler) handleEditMonthCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int) {
	// Parse year and month from callback data: edit_m_YYYY_M
	prefix := callbackEditMonthPrefix
	data := strings.TrimPrefix(q.Data, prefix)
	parts := strings.Split(data, "_")
	if len(parts) != 2 {
		h.answerCallback(q.ID, "Некорректные данные")
		return
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		h.answerCallback(q.ID, "Некорректный год")
		return
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		h.answerCallback(q.ID, "Некорректный месяц")
		return
	}

	// Update state
	state, exists := h.state.GetEditState(userID)
	if !exists {
		h.answerCallback(q.ID, "Сессия устарела")
		return
	}

	state.SelectedYear = year
	state.SelectedMonth = month
	state.Offset = 0
	h.state.SetEditState(userID, state)
	h.state.SetStep(userID, stepEditSelectPurchase)

	// Load purchases for the period
	h.loadAndShowPurchaseForEdit(ctx, q, userID, chatID, messageID, 0)
}

func (h *Handler) handleEditNavigationCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int, action callbackAction) {
	state, exists := h.state.GetEditState(userID)
	if !exists {
		h.answerCallback(q.ID, "Сессия устарела")
		return
	}

	switch action {
	case callbackEditBackToYears:
		// Go back to year selection
		years := extractUniqueYears(state.AvailablePeriods)
		keyboard := buildYearSelectionKeyboard(years)
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "📅 Выберите год:")
		edit.ReplyMarkup = &keyboard
		h.bot.Send(edit)
		h.state.SetStep(userID, stepEditSelectYear)
		h.answerCallback(q.ID, "")
		return

	case callbackEditCancelExpense:
		// Go back to purchase view
		h.loadAndShowPurchaseForEdit(ctx, q, userID, chatID, messageID, state.Offset)
		h.state.SetStep(userID, stepEditSelectPurchase)
		h.answerCallback(q.ID, "")
		return

	case callbackEditFinish:
		// Finish editing
		h.state.ClearEditState(userID)
		h.state.ClearStep(userID)
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "Редактирование завершено")
		edit.ReplyMarkup = nil
		h.bot.Send(edit)
		h.answerCallback(q.ID, "")
		return
	}

	// Calculate new offset for pagination
	newOffset := state.Offset
	switch action {
	case callbackEditPrev:
		newOffset = max(0, state.Offset-editPurchasesPerPage)
	case callbackEditNext:
		newOffset = state.Offset + editPurchasesPerPage
	case callbackEditJumpPrev:
		newOffset = max(0, state.Offset-editJumpStep)
	case callbackEditJumpNext:
		newOffset = min(state.Total-editPurchasesPerPage, state.Offset+editJumpStep)
	}

	h.loadAndShowPurchaseForEdit(ctx, q, userID, chatID, messageID, newOffset)
}

func (h *Handler) loadAndShowPurchaseForEdit(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int, offset int) {
	state, exists := h.state.GetEditState(userID)
	if !exists {
		h.answerCallback(q.ID, "Сессия устарела")
		return
	}

	// Load purchases for the period
	purchasesOut, err := h.useCases.GetPurchasesByPeriod.Execute(ctx, usecase.GetPurchasesByPeriodUseCaseRequest{
		UserID: userID,
		Year:   state.SelectedYear,
		Month:  state.SelectedMonth,
		Limit:  editPurchasesPerPage,
		Offset: offset,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get purchases by period", "error", err)
		h.answerCallback(q.ID, "Ошибка загрузки данных")
		return
	}

	if len(purchasesOut.Purchases) == 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "📭 Нет покупок за выбранный период")
		edit.ReplyMarkup = nil
		h.bot.Send(edit)
		h.answerCallback(q.ID, "")
		return
	}

	// Update state
	state.Offset = offset
	state.Total = purchasesOut.Total
	state.PurchaseID = purchasesOut.Purchases[0].ID
	h.state.SetEditState(userID, state)

	// Show purchase with navigation
	purchase := purchasesOut.Purchases[0]
	message := buildEditPurchaseMessage(purchase, offset, purchasesOut.Total)
	keyboard := buildEditPurchaseKeyboard(offset, purchasesOut.Total, purchase.ID)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, message)
	edit.ReplyMarkup = &keyboard
	h.bot.Send(edit)
	h.answerCallback(q.ID, "")
}

func (h *Handler) handleEditSelectPurchaseCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int) {
	// Parse purchase ID from callback data
	prefix := callbackEditSelectPurchase
	purchaseIDStr := strings.TrimPrefix(q.Data, prefix)
	purchaseID, err := valueobject.NewUUID(purchaseIDStr)
	if err != nil {
		h.answerCallback(q.ID, "Некорректный ID покупки")
		return
	}

	// Get state
	state, exists := h.state.GetEditState(userID)
	if !exists {
		h.answerCallback(q.ID, "Сессия устарела")
		return
	}

	// Load the purchase to get expenses
	purchasesOut, err := h.useCases.GetPurchasesByPeriod.Execute(ctx, usecase.GetPurchasesByPeriodUseCaseRequest{
		UserID: userID,
		Year:   state.SelectedYear,
		Month:  state.SelectedMonth,
		Limit:  editPurchasesPerPage,
		Offset: state.Offset,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get purchase", "error", err)
		h.answerCallback(q.ID, "Ошибка загрузки данных")
		return
	}

	var purchase *entity.Purchase
	for i := range purchasesOut.Purchases {
		if purchasesOut.Purchases[i].ID == purchaseID {
			purchase = &purchasesOut.Purchases[i]
			break
		}
	}

	if purchase == nil {
		h.answerCallback(q.ID, "Покупка не найдена")
		return
	}

	expenses := purchase.Expenses()
	if len(expenses) == 0 {
		h.answerCallback(q.ID, "В покупке нет затрат")
		return
	}

	// Update state
	state.PurchaseID = purchaseID
	h.state.SetEditState(userID, state)
	h.state.SetStep(userID, stepEditSelectExpense)

	// Show expense selection keyboard
	keyboard := buildExpenseSelectionKeyboard(expenses)
	edit := tgbotapi.NewEditMessageText(chatID, messageID, "📝 Выберите затрату для редактирования:")
	edit.ReplyMarkup = &keyboard
	h.bot.Send(edit)
	h.answerCallback(q.ID, "")
}

func (h *Handler) handleEditSelectExpenseCallback(ctx context.Context, q *tgbotapi.CallbackQuery, userID, chatID int64, messageID int) {
	// Parse expense ID from callback data
	prefix := callbackEditSelectExpense
	expenseIDStr := strings.TrimPrefix(q.Data, prefix)
	expenseID, err := valueobject.NewUUID(expenseIDStr)
	if err != nil {
		h.answerCallback(q.ID, "Некорректный ID затраты")
		return
	}

	// Get state
	state, exists := h.state.GetEditState(userID)
	if !exists {
		h.answerCallback(q.ID, "Сессия устарела")
		return
	}

	// Load purchase to get expense details
	purchasesOut, err := h.useCases.GetPurchasesByPeriod.Execute(ctx, usecase.GetPurchasesByPeriodUseCaseRequest{
		UserID: userID,
		Year:   state.SelectedYear,
		Month:  state.SelectedMonth,
		Limit:  editPurchasesPerPage,
		Offset: state.Offset,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get purchase", "error", err)
		h.answerCallback(q.ID, "Ошибка загрузки данных")
		return
	}

	var purchase *entity.Purchase
	for i := range purchasesOut.Purchases {
		if purchasesOut.Purchases[i].ID == state.PurchaseID {
			purchase = &purchasesOut.Purchases[i]
			break
		}
	}

	if purchase == nil {
		h.answerCallback(q.ID, "Покупка не найдена")
		return
	}

	// Find the selected expense
	var selectedExpense *entity.Expense
	for _, exp := range purchase.Expenses() {
		if exp.ID == expenseID {
			selectedExpense = &exp
			break
		}
	}

	if selectedExpense == nil {
		h.answerCallback(q.ID, "Затрата не найдена")
		return
	}

	// Update state with original expense values
	state.ExpenseID = expenseID
	state.OriginalExpenseName = selectedExpense.Name
	state.NewName = selectedExpense.Name
	state.NewQuantity = selectedExpense.Quantity
	state.OriginalUnitPrice = selectedExpense.UnitPrice.DecimalString()
	h.state.SetEditState(userID, state)
	h.state.SetStep(userID, stepEditName)

	// Edit message to remove inline keyboard and show prompt
	msgText := fmt.Sprintf("Редактирование позиции: %s\n\nВведите новые значения (или нажмите кнопку для сохранения старых)", selectedExpense.Name)
	edit := tgbotapi.NewEditMessageText(chatID, messageID, msgText)
	edit.ReplyMarkup = nil
	h.bot.Send(edit)

	// Send new message with reply keyboard for name input
	replyKeyboard := buildEditReplyKeyboard(selectedExpense.Name)
	msg := tgbotapi.NewMessage(chatID, "Введите название:")
	msg.ReplyMarkup = replyKeyboard
	sent, err := h.bot.Send(msg)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to send name prompt", "error", err)
		h.answerCallback(q.ID, "Ошибка")
		return
	}

	// Update message ID in state
	state.MessageID = sent.MessageID
	h.state.SetEditState(userID, state)

	h.answerCallback(q.ID, "")
}

func (h *Handler) answerCallback(callbackID, text string) {
	cfg := tgbotapi.NewCallback(callbackID, text)
	h.bot.Request(cfg)
}
