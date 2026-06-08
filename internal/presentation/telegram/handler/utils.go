package handler

import (
	"context"
	"fms-project/internal/domain/entity"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) sendMessage(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	h.bot.Send(msg)
}

func (h *Handler) sendMessageWithRemoveKeyboard(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = removeReplyKeyboard()
	h.bot.Send(msg)
}

func (h *Handler) downloadLargestPhoto(ctx context.Context, photos []tgbotapi.PhotoSize) ([]byte, error) {
	if len(photos) == 0 {
		return nil, fmt.Errorf("photos are empty")
	}

	photo := photos[len(photos)-1]
	url, err := h.bot.GetFileDirectURL(photo.FileID)
	if err != nil {
		return nil, err
	}

	return h.httpClient.Get(ctx, url)
}

func normalizeDecimalSeparator(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
}

func isOrganisationExists(orgName string, organisations []entity.Organisation) bool {
	searchName := strings.ToLower(orgName)
	for _, org := range organisations {
		if strings.ToLower(strings.TrimSpace(org.Name)) == searchName {
			return true
		}
	}
	return false
}
