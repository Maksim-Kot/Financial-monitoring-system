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

func formatQuantity(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
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

func removeReplyKeyboard() tgbotapi.ReplyKeyboardRemove {
	return tgbotapi.NewRemoveKeyboard(true)
}

// ==================== Format Helpers ====================

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
