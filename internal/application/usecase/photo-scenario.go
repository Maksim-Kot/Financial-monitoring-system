package usecase

import (
	"context"

	"fms-project/internal/application/service"
	"fms-project/internal/application/shared"
	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/gateway"
	"fms-project/internal/domain/repository"
	domainError "fms-project/internal/domain/shared/domain-error"
	"fms-project/internal/infrastructure/logger"
)

type PhotoScenarioUseCaseRequest struct {
	UserID int64
	Photo  []byte
	Text   string
}

type PhotoScenarioUseCaseResponse struct {
}

type PhotoScenarioUseCaseConfig struct {
	Logger             logger.Logger
	StateRepository    repository.UserStateRepository
	PhotoParserGateway gateway.PhotoParserGateway
	ImagePreprocessor  *service.ImagePreprocessorService
}

type PhotoScenarioUseCase struct {
	logger             logger.Logger
	stateRepository    repository.UserStateRepository
	photoParserGateway gateway.PhotoParserGateway
	imagePreprocessor  *service.ImagePreprocessorService
}

func NewPhotoScenarioUseCase(cfg *PhotoScenarioUseCaseConfig) shared.UseCase[PhotoScenarioUseCaseRequest, PhotoScenarioUseCaseResponse] {
	return &PhotoScenarioUseCase{
		logger:             cfg.Logger.With("layer", "usecase", "usecase", "PhotoScenario"),
		stateRepository:    cfg.StateRepository,
		photoParserGateway: cfg.PhotoParserGateway,
		imagePreprocessor:  cfg.ImagePreprocessor,
	}
}

func (uc *PhotoScenarioUseCase) Execute(ctx context.Context, in PhotoScenarioUseCaseRequest) (PhotoScenarioUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID)

	if len(in.Photo) == 0 {
		return PhotoScenarioUseCaseResponse{}, domainError.New(domainError.StatusValidation, "photo is empty")
	}

	processed, err := uc.imagePreprocessor.Process(in.Photo)
	if err != nil {
		logger.ErrorContext(ctx, "failed to preprocess photo", "error", err)
		return PhotoScenarioUseCaseResponse{}, domainError.New(domainError.StatusValidation, "failed to preprocess photo")
	}

	parseOut, err := uc.photoParserGateway.ParsePhoto(ctx, gateway.PhotoParserGatewayIn{Photo: processed})
	if err != nil {
		logger.ErrorContext(ctx, "failed to parse photo", "error", err)
		return PhotoScenarioUseCaseResponse{}, domainError.New(domainError.StatusValidation, "failed to parse photo")
	}
	if len(parseOut.Expenses) == 0 {
		return PhotoScenarioUseCaseResponse{}, domainError.New(domainError.StatusNotFound, "no expenses found")
	}

	logger.DebugContext(ctx, "parsed items from photo", "count", len(parseOut.Expenses))

	state := &entity.UserState{
		UserID:     in.UserID,
		DraftItems: parseOut.Expenses,
		SourceType: entity.SourceTypePhoto,
	}

	if _, err := uc.stateRepository.Save(ctx, repository.UserStateRepositorySaveIn{UserState: state}); err != nil {
		logger.ErrorContext(ctx, "failed to save state", "error", err)
		return PhotoScenarioUseCaseResponse{}, domainError.New(domainError.StatusInternal, "failed to save state")
	}

	return PhotoScenarioUseCaseResponse{}, nil
}
