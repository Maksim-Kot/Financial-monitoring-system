package handler

import (
	"fmt"
	"strconv"
	"strings"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/valueobject"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func buildPurchasesMessage(purchases []entity.Purchase, offset, total int) string {
	if len(purchases) == 0 {
		return "📋 Покупок пока нет"
	}

	var b strings.Builder
	pageNum := offset/purchasesPerPage + 1
	totalPages := (total + purchasesPerPage - 1) / purchasesPerPage
	if totalPages == 0 {
		totalPages = 1
	}
	fmt.Fprintf(&b, "📋 Покупки (страница %d из %d):\n\n", pageNum, totalPages)

	for i, purchase := range purchases {
		itemNum := offset + i + 1
		orgName := strings.TrimSpace(purchase.OrganisationName)
		if orgName == "" {
			orgName = "не указано"
		}

		fmt.Fprintf(&b, "%d) %s 🏪\n", itemNum, orgName)
		fmt.Fprintf(&b, "Дата: %s\n", purchase.PurchaseDate.Format("02.01.2006"))

		expenses := purchase.Expenses()
		if len(expenses) > 0 {
			b.WriteString("🛒 Товары:\n")
			for _, expense := range expenses {
				total, err := expense.TotalPrice()
				priceStr := "?"
				if err == nil {
					priceStr = total.DecimalString()
				}

				qty := formatQuantity(expense.Quantity)
				name := strings.TrimSpace(expense.Name)
				if name == "" {
					name = "неизвестно"
				}

				fmt.Fprintf(&b, "   • %s | %s шт. | %s BYN\n", name, qty, priceStr)
			}
		}

		total, err := purchase.TotalPrice()
		if err == nil {
			fmt.Fprintf(&b, "Итого: %s BYN\n", total.DecimalString())
		}

		if i < len(purchases)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func buildListKeyboard(offset, total int) tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.InlineKeyboardButton

	if offset > 0 {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", string(callbackListPrev)))
	}

	buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Завершить", string(callbackListFinish)))

	if offset+purchasesPerPage < total {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Вперед ▶️", string(callbackListNext)))
	}

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttons...),
	)
}

func buildOrganisationsReplyKeyboard(organisations []entity.Organisation) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton

	for i := 0; i < len(organisations); i += 2 {
		if i+1 < len(organisations) {
			rows = append(rows, []tgbotapi.KeyboardButton{
				tgbotapi.NewKeyboardButton(organisations[i].Name),
				tgbotapi.NewKeyboardButton(organisations[i+1].Name),
			})
		} else {
			rows = append(rows, []tgbotapi.KeyboardButton{
				tgbotapi.NewKeyboardButton(organisations[i].Name),
			})
		}
	}

	return tgbotapi.NewReplyKeyboard(rows...)
}

func buildSaveOrgKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", string(callbackSaveOrgYes)),
			tgbotapi.NewInlineKeyboardButtonData("Нет", string(callbackSaveOrgNo)),
		),
	)
}

func buildAddMoreKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", string(callbackAddMoreYes)),
			tgbotapi.NewInlineKeyboardButtonData("Нет", string(callbackAddMoreNo)),
		),
	)
}

// ==================== Analytics View Builders ====================

func buildPeriodSelectionKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 День", string(callbackStatsDay)),
			tgbotapi.NewInlineKeyboardButtonData("📅 Неделя", string(callbackStatsWeek)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Месяц", string(callbackStatsMonth)),
			tgbotapi.NewInlineKeyboardButtonData("📅 Полгода", string(callbackStatsHalfYear)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", string(callbackStatsClose)),
		),
	)
}

func buildSummaryMessage(summary entity.Summary) string {
	var b strings.Builder

	fmt.Fprintf(&b, "📊 Аналитика: %s\n\n", summary.Period.Format())
	fmt.Fprintf(&b, "💰 Общие расходы: %s\n", formatMoney(summary.Total))
	fmt.Fprintf(&b, "   %d покупок • %d позиций\n\n", summary.PurchaseCount, summary.ExpenseCount)

	if len(summary.TopCategories) > 0 {
		b.WriteString("🏆 Топ категорий:\n")
		for _, cat := range summary.TopCategories {
			icon := cat.CategoryIcon
			if icon == "" {
				icon = "📦"
			}
			fmt.Fprintf(&b, "   %s %s — %s (%s)\n",
				icon, cat.CategoryName, formatMoney(cat.Total), formatPercentage(cat.Percentage))
		}
		b.WriteString("\n")
	}

	if len(summary.TopPurchases) > 0 {
		b.WriteString("💎 Самые крупные покупки:\n")
		for i, p := range summary.TopPurchases {
			fmt.Fprintf(&b, "   %d. %s — %s (%s)\n",
				i+1, p.OrganisationName,
				formatMoney(p.Total), p.PurchaseDate.Format("02.01"))
		}
	}

	return b.String()
}

func buildDetailedMessage(report entity.DetailedReport) string {
	var b strings.Builder

	fmt.Fprintf(&b, "📊 Детальный анализ: %s\n\n", report.Summary.Period.Format())

	// Comparison section
	fmt.Fprintf(&b, "📈 Сравнение с %s:\n", report.Comparison.PreviousPeriod.Format())
	fmt.Fprintf(&b, "   Текущий: %s\n", formatMoney(report.Comparison.CurrentTotal))
	fmt.Fprintf(&b, "   Предыдущий: %s\n", formatMoney(report.Comparison.PreviousTotal))
	fmt.Fprintf(&b, "   Изменение: %s\n\n", formatDelta(report.Comparison.DeltaPercent))

	// Category deltas
	increases, decreases := separateDeltas(report.CategoryDeltas)

	if len(increases) > 0 {
		b.WriteString("📊 Что выросло:\n")
		for _, d := range increases[:min(3, len(increases))] {
			fmt.Fprintf(&b, "   • %s %s (%s → %s)\n",
				d.CategoryName, formatDelta(d.DeltaPercent),
				formatMoney(d.PreviousTotal), formatMoney(d.CurrentTotal))
		}
		b.WriteString("\n")
	}

	if len(decreases) > 0 {
		b.WriteString("📉 Что сократилось:\n")
		for _, d := range decreases[:min(3, len(decreases))] {
			fmt.Fprintf(&b, "   • %s %s (%s → %s)\n",
				d.CategoryName, formatDelta(d.DeltaPercent),
				formatMoney(d.PreviousTotal), formatMoney(d.CurrentTotal))
		}
		b.WriteString("\n")
	}

	// Anomalies
	if len(report.Anomalies) > 0 {
		fmt.Fprintf(&b, "🔍 Аномальные покупки (%d):\n", len(report.Anomalies))
		for i, a := range report.Anomalies[:min(3, len(report.Anomalies))] {
			fmt.Fprintf(&b, "   %d. %s — %s\n", i+1, a.Name, formatMoney(a.Total))
			fmt.Fprintf(&b, "      (в %.1fx раз выше среднего)\n", a.Factor)
		}
		b.WriteString("\n")
	}

	// AI Insights
	if len(report.Insights) > 0 {
		b.WriteString("💡 Инсайты:\n")
		for _, insight := range report.Insights {
			fmt.Fprintf(&b, "   • %s\n", insight.Text)
		}
	}

	return b.String()
}

func buildDetailedPromptKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Подробный анализ", string(callbackStatsDetailedYes)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Закрыть", string(callbackStatsClose)),
		),
	)
}

func buildDetailedNavigationKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("◀️ К сводке", string(callbackStatsBack)),
			tgbotapi.NewInlineKeyboardButtonData("📅 Другой период", string(callbackStatsOtherPeriod)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Закрыть", string(callbackStatsClose)),
		),
	)
}

func buildNoDataKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Другой период", string(callbackStatsOtherPeriod)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Закрыть", string(callbackStatsClose)),
		),
	)
}

// ==================== Edit Scenario View Builders ====================

func buildYearSelectionKeyboard(years []int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Create rows with 3 buttons each
	for i := 0; i < len(years); i += 3 {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < 3 && i+j < len(years); j++ {
			year := years[i+j]
			callbackData := fmt.Sprintf("%s%d", callbackEditYearPrefix, year)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				strconv.Itoa(year),
				callbackData,
			))
		}
		rows = append(rows, row)
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", string(callbackEditFinish)),
	})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func buildMonthSelectionKeyboard(year int, months []int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Create rows with 3 buttons each
	for i := 0; i < len(months); i += 3 {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < 3 && i+j < len(months); j++ {
			month := months[i+j]
			monthName := monthNames[month]
			if monthName == "" {
				monthName = strconv.Itoa(month)
			}
			callbackData := fmt.Sprintf("%s%d_%d", callbackEditMonthPrefix, year, month)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				monthName,
				callbackData,
			))
		}
		rows = append(rows, row)
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("◀️ К годам", string(callbackEditBackToYears)),
	})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func buildEditPurchaseMessage(purchase entity.Purchase, offset, total int) string {
	var b strings.Builder

	pageNum := offset/editPurchasesPerPage + 1
	totalPages := (total + editPurchasesPerPage - 1) / editPurchasesPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	orgName := strings.TrimSpace(purchase.OrganisationName)
	if orgName == "" {
		orgName = "не указано"
	}

	fmt.Fprintf(&b, "📝 Покупка %d из %d:\n\n", pageNum, totalPages)
	fmt.Fprintf(&b, "🏪 %s\n", orgName)
	fmt.Fprintf(&b, "Дата: %s\n\n", purchase.PurchaseDate.Format("02.01.2006"))

	expenses := purchase.Expenses()
	if len(expenses) > 0 {
		b.WriteString("🛒 Позиции:\n")
		for i, expense := range expenses {
			total, err := expense.TotalPrice()
			priceStr := "?"
			if err == nil {
				priceStr = total.DecimalString()
			}

			qty := formatQuantity(expense.Quantity)
			name := strings.TrimSpace(expense.Name)
			if name == "" {
				name = "неизвестно"
			}

			fmt.Fprintf(&b, "%d. %s | %s шт. | %s BYN\n", i+1, name, qty, priceStr)
		}
		b.WriteString("\n")
	}

	totalPrice, err := purchase.TotalPrice()
	if err == nil {
		fmt.Fprintf(&b, "💰 Итого: %s BYN", totalPrice.DecimalString())
	}

	return b.String()
}

func buildEditPurchaseKeyboard(offset, total int, purchaseID valueobject.UUID) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Navigation row: Prev, Edit, Next
	var navRow []tgbotapi.InlineKeyboardButton

	if offset > 0 {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData(
			"◀️ Назад",
			string(callbackEditPrev),
		))
	}

	// Edit button with purchase ID
	editCallback := fmt.Sprintf("%s%s", callbackEditSelectPurchase, purchaseID.String())
	navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData(
		"✏️ Изменить",
		editCallback,
	))

	if offset+editPurchasesPerPage < total {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData(
			"Вперед ▶️",
			string(callbackEditNext),
		))
	}

	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}

	// Jump row: -5, Finish, +5
	var jumpRow []tgbotapi.InlineKeyboardButton

	if offset >= editJumpStep {
		jumpRow = append(jumpRow, tgbotapi.NewInlineKeyboardButtonData(
			"⏪ -5",
			string(callbackEditJumpPrev),
		))
	}

	jumpRow = append(jumpRow, tgbotapi.NewInlineKeyboardButtonData(
		"✅ Завершить",
		string(callbackEditFinish),
	))

	if offset+editPurchasesPerPage+editJumpStep <= total {
		jumpRow = append(jumpRow, tgbotapi.NewInlineKeyboardButtonData(
			"⏩ +5",
			string(callbackEditJumpNext),
		))
	}

	if len(jumpRow) > 0 {
		rows = append(rows, jumpRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func buildExpenseSelectionKeyboard(expenses []entity.Expense) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Create a button for each expense
	for _, expense := range expenses {
		name := strings.TrimSpace(expense.Name)
		if name == "" {
			name = "неизвестно"
		}
		callbackData := fmt.Sprintf("%s%s", callbackEditSelectExpense, expense.ID.String())
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(name, callbackData),
		})
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", string(callbackEditCancelExpense)),
	})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func buildEditReplyKeyboard(currentValue string) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(currentValue),
		),
	)
}

func extractUniqueYears(periods []PeriodInfo) []int {
	years := make([]int, len(periods))
	for i, p := range periods {
		years[i] = p.Year
	}
	return years
}

func removeReplyKeyboard() tgbotapi.ReplyKeyboardRemove {
	return tgbotapi.NewRemoveKeyboard(true)
}

// ==================== Format Helpers ====================

func formatQuantity(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatMoney(amount valueobject.MoneyAmount) string {
	return fmt.Sprintf("%s BYN", amount.DecimalString())
}

func formatDelta(delta float64) string {
	if delta > 0 {
		return fmt.Sprintf("📈 +%.1f%%", delta)
	}
	if delta < 0 {
		return fmt.Sprintf("📉 %.1f%%", delta)
	}
	return "➡️ 0%"
}

func formatPercentage(p float64) string {
	return fmt.Sprintf("%.1f%%", p)
}

func separateDeltas(deltas []entity.CategoryDelta) ([]entity.CategoryDelta, []entity.CategoryDelta) {
	increases := make([]entity.CategoryDelta, 0)
	decreases := make([]entity.CategoryDelta, 0)

	for _, d := range deltas {
		if d.DeltaPercent > 0 {
			increases = append(increases, d)
		} else if d.DeltaPercent < 0 {
			decreases = append(decreases, d)
		}
	}

	return increases, decreases
}

var monthNames = map[int]string{
	1:  "Январь",
	2:  "Февраль",
	3:  "Март",
	4:  "Апрель",
	5:  "Май",
	6:  "Июнь",
	7:  "Июль",
	8:  "Август",
	9:  "Сентябрь",
	10: "Октябрь",
	11: "Ноябрь",
	12: "Декабрь",
}
