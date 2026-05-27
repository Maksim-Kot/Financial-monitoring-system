package handler

import (
	"context"
	"fmt"
	"strings"

	"fms-project/internal/application/usecase"
	domainError "fms-project/internal/domain/shared/domain-error"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleStep(ctx context.Context, userID, chatID int64, msg *tgbotapi.Message) {
	switch h.state.GetStep(userID) {
	case stepWaitText:
		h.stepText(ctx, userID, chatID, msg.Text)
	case stepWaitPhoto:
		if len(msg.Photo) == 0 {
			h.sendMessage(chatID, "Ожидаю фото. Отправьте изображение чека или используйте /cancel для отмены")
			return
		}
		h.stepPhoto(ctx, userID, chatID, msg.Photo)
	case stepWaitManualItem:
		h.stepManualItem(ctx, userID, chatID, msg.Text)
	case stepWaitDate:
		h.stepDate(ctx, userID, chatID, msg.Text)
	case stepWaitOrganisation:
		h.stepOrganisation(ctx, userID, chatID, msg.Text)
	default:
		h.sendMessage(chatID, "Используйте команды:\n/start - приветствие\n/text - добавить покупку текстом\n/photo - добавить покупку по фото\n/cancel - отмена")
	}
}

func (h *Handler) stepText(ctx context.Context, userID, chatID int64, text string) {
	if strings.TrimSpace(text) == "" {
		h.sendMessage(chatID, "Текст не может быть пустым. Попробуйте еще раз или используйте /cancel")
		return
	}

	_, err := h.useCases.TextScenario.Execute(ctx, usecase.TextScenarioUseCaseRequest{
		UserID: userID,
		Text:   text,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to execute text scenario", "error", err)
		h.sendMessage(chatID, "Не удалось обработать текст. Попробуйте еще раз или используйте /cancel")
		return
	}

	_, err = h.useCases.ClassifyCategory.Execute(ctx, usecase.ClassifyCategoryUseCaseRequest{
		UserID: userID,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to classify categories", "error", err)
		h.sendMessage(chatID, "Не удалось обработать текст. Попробуйте еще раз или используйте /cancel")
		return
	}

	h.state.SetStep(userID, stepWaitDate)
	h.sendMessage(chatID, "Теперь введите дату покупки в формате ДД.ММ.ГГГГ")
}

func (h *Handler) stepPhoto(ctx context.Context, userID, chatID int64, photos []tgbotapi.PhotoSize) {
	msg, _ := h.bot.Send(tgbotapi.NewMessage(chatID, "Обрабатываю фото..."))

	photoData, err := h.downloadLargestPhoto(ctx, photos)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to download photo", "error", err)
		h.sendMessage(chatID, "Не удалось загрузить фото. Попробуйте еще раз или используйте /cancel")
		return
	}

	_, err = h.useCases.PhotoScenario.Execute(ctx, usecase.PhotoScenarioUseCaseRequest{
		UserID: userID,
		Photo:  photoData,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to execute photo scenario", "error", err)
		h.sendMessage(chatID, "Не удалось распознать чек. Попробуйте другое фото или используйте /cancel")
		return
	}

	_, err = h.useCases.ClassifyCategory.Execute(ctx, usecase.ClassifyCategoryUseCaseRequest{
		UserID: userID,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to classify categories", "error", err)
		h.sendMessage(chatID, "Не удалось обработать фото. Попробуйте еще раз или используйте /cancel")
		return
	}

	h.state.SetStep(userID, stepWaitDate)

	confirmMsg := "Фото обработано! Теперь введите дату покупки в формате ДД.ММ.ГГГГ"
	edit := tgbotapi.NewEditMessageText(chatID, msg.MessageID, confirmMsg)
	h.bot.Send(edit)
}

func (h *Handler) stepManualItem(ctx context.Context, userID, chatID int64, text string) {
	if strings.TrimSpace(text) == "" {
		h.sendMessage(chatID, "Ввод не может быть пустым. Попробуйте еще раз или используйте /cancel")
		return
	}

	out, err := h.useCases.ManualAddItem.Execute(ctx, usecase.ManualAddItemUseCaseRequest{
		UserID: userID,
		Item:   text,
	})
	if err != nil {
		if domainError.HasStatus(err, domainError.StatusValidation) {
			h.sendMessage(chatID, "Неверный формат. Введите позицию в формате:\nнаименование; количество; цена за единицу\n\nНапример: Молоко; 2; 1.50")
			return
		}
		h.logger.ErrorContext(ctx, "failed to add manual item", "error", err)
		h.sendMessage(chatID, "Произошла ошибка. Попробуйте еще раз или используйте /cancel")
		return
	}

	total := out.Item.UnitPrice * out.Item.Quantity
	msg := fmt.Sprintf("Добавлена позиция к покупке:\n%s — %.2f шт. × %.2f = %.2f BYN\n\nДобавить еще позицию?", out.Item.Name, out.Item.Quantity, out.Item.UnitPrice, total)

	msgCfg := tgbotapi.NewMessage(chatID, msg)
	msgCfg.ReplyMarkup = buildAddMoreKeyboard()

	sentMsg, err := h.bot.Send(msgCfg)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to send add more message", "error", err)
		h.sendMessage(chatID, "Произошла ошибка. Используйте /cancel для отмены")
		return
	}

	h.state.SetStep(userID, stepWaitAddMoreItem)
	h.state.SetManualItemsData(userID, manualItemsData{
		MessageID: sentMsg.MessageID,
	})
}

func (h *Handler) stepDate(ctx context.Context, userID, chatID int64, text string) {
	_, err := h.useCases.AddPurchaseDate.Execute(ctx, usecase.AddPurchaseDateUseCaseRequest{
		UserID: userID,
		Date:   text,
	})
	if err != nil {
		if domainError.HasStatus(err, domainError.StatusValidation) {
			h.sendMessage(chatID, "Неверный формат даты. Введите дату в формате ДД.ММ.ГГГГ (например, 15.01.2024)")
			return
		}
		h.logger.ErrorContext(ctx, "failed to add purchase date", "error", err)
		h.sendMessage(chatID, "Произошла ошибка. Попробуйте еще раз или используйте /cancel")
		return
	}

	listOut, err := h.useCases.ListOrganisations.Execute(ctx, usecase.ListOrganisationsUseCaseRequest{
		UserID: userID,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to list user organisations", "error", err)
		h.sendMessage(chatID, "Произошла ошибка. Попробуйте еще раз или используйте /cancel")
		return
	}

	h.state.SetStep(userID, stepWaitOrganisation)

	if len(listOut.Organisations) > 0 {
		msg := tgbotapi.NewMessage(chatID, "Выберите организацию (магазин) из списка или введите новую:")
		msg.ReplyMarkup = buildOrganisationsReplyKeyboard(listOut.Organisations)
		h.bot.Send(msg)
	} else {
		h.sendMessage(chatID, "Введите название организации (магазина):")
	}
}

func (h *Handler) stepOrganisation(ctx context.Context, userID, chatID int64, text string) {
	orgName := strings.TrimSpace(text)
	if orgName == "" {
		h.sendMessage(chatID, "Название организации не может быть пустым. Попробуйте еще раз или используйте /cancel")
		return
	}

	_, err := h.useCases.AddOrganisation.Execute(ctx, usecase.AddOrganisationUseCaseRequest{
		UserID: userID,
		Name:   orgName,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to add organisation", "error", err)
		h.sendMessage(chatID, "Произошла ошибка. Попробуйте еще раз или используйте /cancel")
		return
	}

	listOut, err := h.useCases.ListOrganisations.Execute(ctx, usecase.ListOrganisationsUseCaseRequest{
		UserID: userID,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to list user organisations", "error", err)
		h.finalizePurchase(ctx, userID, chatID)
		return
	}

	if len(listOut.Organisations) >= maxSavedOrganisations {
		h.sendMessageWithRemoveKeyboard(chatID, "Достигнут лимит сохраненных организаций.")
		h.finalizePurchase(ctx, userID, chatID)
		return
	}

	exists := isOrganisationExists(orgName, listOut.Organisations)
	if exists {
		h.finalizePurchase(ctx, userID, chatID)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Сохранить организацию \""+orgName+"\" для быстрого выбора в будущем?")
	msg.ReplyMarkup = buildSaveOrgKeyboard()

	sentMsg, err := h.bot.Send(msg)
	if err != nil {
		h.finalizePurchase(ctx, userID, chatID)
		return
	}

	h.state.SetSaveOrgData(userID, saveOrgData{
		MessageID: sentMsg.MessageID,
		Name:      orgName,
	})
	h.state.SetStep(userID, stepWaitSaveOrg)
}

func (h *Handler) finalizePurchase(ctx context.Context, userID, chatID int64) {
	_, err := h.useCases.FinalizeScenario.Execute(ctx, usecase.FinalizeScenarioUseCaseRequest{
		UserID: userID,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to finalize scenario", "error", err)
		h.sendMessageWithRemoveKeyboard(chatID, "Произошла ошибка при сохранении. Попробуйте позже")
		return
	}

	h.state.ClearStep(userID)
	h.state.ClearManualItemsData(userID)
	h.sendMessageWithRemoveKeyboard(chatID, "Покупка успешно сохранена!")
}
