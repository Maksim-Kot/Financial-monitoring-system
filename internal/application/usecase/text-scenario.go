package usecase

import (
	"context"
	"strings"

	"fms-project/internal/application/shared"
	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/gateway"
	"fms-project/internal/domain/repository"
	domainError "fms-project/internal/domain/shared/domain-error"
	"fms-project/internal/infrastructure/logger"
)

type TextScenarioUseCaseRequest struct {
	UserID int64
	Text   string
}

type TextScenarioUseCaseResponse struct {
}

type TextScenarioUseCaseConfig struct {
	Logger            logger.Logger
	StateRepository   repository.UserStateRepository
	TextParserGateway gateway.TextParserGateway
}

type TextScenarioUseCase struct {
	logger            logger.Logger
	stateRepository   repository.UserStateRepository
	textParserGateway gateway.TextParserGateway
}

func NewTextScenarioUseCase(cfg *TextScenarioUseCaseConfig) shared.UseCase[TextScenarioUseCaseRequest, TextScenarioUseCaseResponse] {
	return &TextScenarioUseCase{
		logger:            cfg.Logger.With("layer", "usecase", "usecase", "TextScenario"),
		stateRepository:   cfg.StateRepository,
		textParserGateway: cfg.TextParserGateway,
	}
}

func (uc *TextScenarioUseCase) Execute(ctx context.Context, in TextScenarioUseCaseRequest) (TextScenarioUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID)

	if strings.TrimSpace(in.Text) == "" {
		return TextScenarioUseCaseResponse{}, domainError.New(domainError.StatusValidation, "text is empty")
	}

	parseOut, err := uc.textParserGateway.ParseText(ctx, gateway.TextParserGatewayIn{Text: in.Text})
	if err != nil {
		logger.ErrorContext(ctx, "failed to parse text", "error", err)
		return TextScenarioUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to parse text")
	}
	if len(parseOut.Expenses) == 0 {
		return TextScenarioUseCaseResponse{}, domainError.New(domainError.StatusNotFound, "no expenses found")
	}

	logger.DebugContext(ctx, "parsed items from text", "count", len(parseOut.Expenses))

	state := &entity.UserState{
		UserID:     in.UserID,
		DraftItems: parseOut.Expenses,
		SourceType: entity.SourceTypeText,
	}

	if _, err := uc.stateRepository.Save(ctx, repository.UserStateRepositorySaveIn{UserState: state}); err != nil {
		logger.ErrorContext(ctx, "failed to save state", "error", err)
		return TextScenarioUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to save state")
	}

	return TextScenarioUseCaseResponse{}, nil
}
