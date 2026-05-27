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
