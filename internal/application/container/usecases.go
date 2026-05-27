package container

import (
	"fms-project/internal/application/shared"
	"fms-project/internal/application/usecase"
	"fms-project/internal/infrastructure/config"
	"fms-project/internal/infrastructure/logger"
)

type UseCasesContainer struct {
	EnsureCategories shared.UseCase[usecase.EnsureCategoriesUseCaseRequest, usecase.EnsureCategoriesUseCaseResponse]

	TextScenario     shared.UseCase[usecase.TextScenarioUseCaseRequest, usecase.TextScenarioUseCaseResponse]
	PhotoScenario    shared.UseCase[usecase.PhotoScenarioUseCaseRequest, usecase.PhotoScenarioUseCaseResponse]
	ManualAddItem    shared.UseCase[usecase.ManualAddItemUseCaseRequest, usecase.ManualAddItemUseCaseResponse]
	ClassifyCategory shared.UseCase[usecase.ClassifyCategoryUseCaseRequest, usecase.ClassifyCategoryUseCaseResponse]
	AddPurchaseDate  shared.UseCase[usecase.AddPurchaseDateUseCaseRequest, usecase.AddPurchaseDateUseCaseResponse]
	AddOrganisation  shared.UseCase[usecase.AddOrganisationUseCaseRequest, usecase.AddOrganisationUseCaseResponse]
	FinalizeScenario shared.UseCase[usecase.FinalizeScenarioUseCaseRequest, usecase.FinalizeScenarioUseCaseResponse]
	ListScenario     shared.UseCase[usecase.ListScenarioUseCaseRequest, usecase.ListScenarioUseCaseResponse]

	ListOrganisations shared.UseCase[usecase.ListOrganisationsUseCaseRequest, usecase.ListOrganisationsUseCaseResponse]
	SaveOrganisation  shared.UseCase[usecase.SaveOrganisationUseCaseRequest, usecase.SaveOrganisationUseCaseResponse]

	GetAnalytics shared.UseCase[usecase.GetAnalyticsUseCaseRequest, usecase.GetAnalyticsUseCaseResponse]

	CancelScenario shared.UseCase[usecase.CancelScenarioUseCaseRequest, usecase.CancelScenarioUseCaseResponse]
}

type UseCasesContainerConfig struct {
	Config       *config.Config
	Logger       logger.Logger
	Services     *ServicesContainer
	Repositories *RepositoriesContainer
	Gateways     *GatewaysContainer
}

func NewUseCasesContainer(cfg *UseCasesContainerConfig) *UseCasesContainer {
	ensureCategoriesUseCase := usecase.NewEnsureCategoriesUseCase(&usecase.EnsureCategoriesUseCaseConfig{
		Logger:             cfg.Logger,
		CategoryRepository: cfg.Repositories.Categories,
		TxManager:          cfg.Repositories.TxManager,
	})

	textScenarioUseCase := usecase.NewTextScenarioUseCase(&usecase.TextScenarioUseCaseConfig{
		Logger:            cfg.Logger,
		StateRepository:   cfg.Repositories.UserState,
		TextParserGateway: cfg.Gateways.TextParser,
	})

	photoScenarioUseCase := usecase.NewPhotoScenarioUseCase(&usecase.PhotoScenarioUseCaseConfig{
		Logger:             cfg.Logger,
		StateRepository:    cfg.Repositories.UserState,
		PhotoParserGateway: cfg.Gateways.PhotoParser,
		ImagePreprocessor:  cfg.Services.ImagePreprocessor,
	})

	manualAddItemUseCase := usecase.NewManualAddItemUseCase(&usecase.ManualAddItemUseCaseConfig{
		Logger:          cfg.Logger,
		StateRepository: cfg.Repositories.UserState,
	})

	classifyCategoryUseCase := usecase.NewClassifyCategoryUseCase(&usecase.ClassifyCategoryUseCaseConfig{
		Logger:                    cfg.Logger,
		StateRepository:           cfg.Repositories.UserState,
		CategoryRepository:        cfg.Repositories.Categories,
		CategoryMatcher:           cfg.Services.CategoryMatcher,
		CategoryClassifierGateway: cfg.Gateways.CategoryClassifier,
	})

	addPurchaseDateUseCase := usecase.NewAddPurchaseDateUseCase(&usecase.AddPurchaseDateUseCaseConfig{
		Logger:          cfg.Logger,
		StateRepository: cfg.Repositories.UserState,
	})

	addOrganisationUseCase := usecase.NewAddOrganisationUseCase(&usecase.AddOrganisationUseCaseConfig{
		Logger:                 cfg.Logger,
		StateRepository:        cfg.Repositories.UserState,
		OrganisationRepository: cfg.Repositories.Organisations,
	})

	finalizeScenarioUseCase := usecase.NewFinalizeScenarioUseCase(&usecase.FinalizeScenarioUseCaseConfig{
		Logger:             cfg.Logger,
		StateRepository:    cfg.Repositories.UserState,
		PurchaseRepository: cfg.Repositories.Purchases,
		TxManager:          cfg.Repositories.TxManager,
	})

	listScenarioUseCase := usecase.NewListScenarioUseCase(&usecase.ListScenarioUseCaseConfig{
		Logger:             cfg.Logger,
		PurchaseRepository: cfg.Repositories.Purchases,
	})

	listOrganisationsUseCase := usecase.NewListOrganisationsUseCase(&usecase.ListOrganisationsUseCaseConfig{
		Logger:                 cfg.Logger,
		OrganisationRepository: cfg.Repositories.Organisations,
	})

	saveOrganisationUseCase := usecase.NewSaveOrganisationUseCase(&usecase.SaveOrganisationUseCaseConfig{
		Logger:                 cfg.Logger,
		OrganisationRepository: cfg.Repositories.Organisations,
		TxManager:              cfg.Repositories.TxManager,
	})

	getAnalyticsUseCase := usecase.NewGetAnalyticsUseCase(&usecase.GetAnalyticsUseCaseConfig{
		Logger:           cfg.Logger,
		AnalyticsRepo:    cfg.Repositories.Analytics,
		Calculator:       cfg.Services.AnalyticsCalculator,
		AnomalyDetector:  cfg.Services.AnomalyDetector,
		InsightGenerator: cfg.Gateways.InsightGenerator,
	})

	cancelScenarioUseCase := usecase.NewCancelScenarioUseCase(&usecase.CancelScenarioUseCaseConfig{
		Logger:          cfg.Logger,
		StateRepository: cfg.Repositories.UserState,
	})

	return &UseCasesContainer{
		EnsureCategories:  ensureCategoriesUseCase,
		TextScenario:      textScenarioUseCase,
		PhotoScenario:     photoScenarioUseCase,
		ManualAddItem:     manualAddItemUseCase,
		ClassifyCategory:  classifyCategoryUseCase,
		AddPurchaseDate:   addPurchaseDateUseCase,
		AddOrganisation:   addOrganisationUseCase,
		FinalizeScenario:  finalizeScenarioUseCase,
		ListScenario:      listScenarioUseCase,
		ListOrganisations: listOrganisationsUseCase,
		SaveOrganisation:  saveOrganisationUseCase,
		GetAnalytics:      getAnalyticsUseCase,
		CancelScenario:    cancelScenarioUseCase,
	}
}
