package handler

import (
	"sync"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/valueobject"
)

type userStep string

const (
	stepIdle             userStep = "idle"
	stepWaitText         userStep = "wait_text"
	stepWaitPhoto        userStep = "wait_photo"
	stepWaitDate         userStep = "wait_date"
	stepWaitOrganisation userStep = "wait_org"
	stepWaitSaveOrg      userStep = "wait_save_org"
	stepWaitManualItem   userStep = "wait_manual_item"
	stepWaitAddMoreItem  userStep = "wait_add_more_item"

	// Edit scenario steps
	stepEditSelectYear     userStep = "edit_select_year"
	stepEditSelectMonth    userStep = "edit_select_month"
	stepEditSelectPurchase userStep = "edit_select_purchase"
	stepEditSelectExpense  userStep = "edit_select_expense"
	stepEditName           userStep = "edit_name"
	stepEditQuantity       userStep = "edit_quantity"
	stepEditPrice          userStep = "edit_price"
)

type listPaginationState struct {
	MessageID int
	Offset    int
	Total     int
}

type saveOrgData struct {
	MessageID int
	Name      string
}

type manualItemsData struct {
	MessageID int
}

type analyticsState struct {
	MessageID  int
	PeriodType valueobject.PeriodType
	Summary    entity.Summary
}

type editState struct {
	// Available periods loaded once at the start
	AvailablePeriods []PeriodInfo
	// Selected period
	SelectedYear  int
	SelectedMonth int
	// Current purchase being viewed
	PurchaseID valueobject.UUID
	Offset     int
	Total      int
	MessageID  int
	// Expense being edited
	ExpenseID valueobject.UUID
	// Intermediate edit values
	NewName     string
	NewQuantity float64
	// Original expense values for reference and keyboard suggestions
	OriginalExpenseName string
	OriginalUnitPrice   string
}

type PeriodInfo struct {
	Year   int
	Months []int
}

type State struct {
	userSteps         map[int64]userStep
	listStates        map[int64]listPaginationState
	saveOrgStates     map[int64]saveOrgData
	manualItemsStates map[int64]manualItemsData
	analyticsStates   map[int64]analyticsState
	editStates        map[int64]editState

	stepsMu       sync.RWMutex
	listMu        sync.RWMutex
	saveOrgMu     sync.RWMutex
	manualItemsMu sync.RWMutex
	analyticsMu   sync.RWMutex
	editMu        sync.RWMutex
}

func NewState() *State {
	return &State{
		userSteps:         make(map[int64]userStep),
		listStates:        make(map[int64]listPaginationState),
		saveOrgStates:     make(map[int64]saveOrgData),
		manualItemsStates: make(map[int64]manualItemsData),
		analyticsStates:   make(map[int64]analyticsState),
		editStates:        make(map[int64]editState),
	}
}

// Steps

func (s *State) GetStep(userID int64) userStep {
	s.stepsMu.RLock()
	defer s.stepsMu.RUnlock()

	step, ok := s.userSteps[userID]
	if !ok {
		return stepIdle
	}
	return step
}

func (s *State) SetStep(userID int64, step userStep) {
	s.stepsMu.Lock()
	defer s.stepsMu.Unlock()
	s.userSteps[userID] = step
}

func (s *State) ClearStep(userID int64) {
	s.stepsMu.Lock()
	defer s.stepsMu.Unlock()
	delete(s.userSteps, userID)
}

// Pagination

func (s *State) GetList(userID int64) (listPaginationState, bool) {
	s.listMu.RLock()
	defer s.listMu.RUnlock()

	st, ok := s.listStates[userID]
	return st, ok
}

func (s *State) SetList(userID int64, st listPaginationState) {
	s.listMu.Lock()
	defer s.listMu.Unlock()
	s.listStates[userID] = st
}

func (s *State) ClearList(userID int64) {
	s.listMu.Lock()
	defer s.listMu.Unlock()
	delete(s.listStates, userID)
}

// Save Organisation Data

func (s *State) GetSaveOrgData(userID int64) (saveOrgData, bool) {
	s.saveOrgMu.RLock()
	defer s.saveOrgMu.RUnlock()

	data, ok := s.saveOrgStates[userID]
	return data, ok
}

func (s *State) SetSaveOrgData(userID int64, data saveOrgData) {
	s.saveOrgMu.Lock()
	defer s.saveOrgMu.Unlock()
	s.saveOrgStates[userID] = data
}

func (s *State) ClearSaveOrgData(userID int64) {
	s.saveOrgMu.Lock()
	defer s.saveOrgMu.Unlock()
	delete(s.saveOrgStates, userID)
}

// Manual Items Data

func (s *State) GetManualItemsData(userID int64) (manualItemsData, bool) {
	s.manualItemsMu.RLock()
	defer s.manualItemsMu.RUnlock()

	data, ok := s.manualItemsStates[userID]
	return data, ok
}

func (s *State) SetManualItemsData(userID int64, data manualItemsData) {
	s.manualItemsMu.Lock()
	defer s.manualItemsMu.Unlock()
	s.manualItemsStates[userID] = data
}

func (s *State) ClearManualItemsData(userID int64) {
	s.manualItemsMu.Lock()
	defer s.manualItemsMu.Unlock()
	delete(s.manualItemsStates, userID)
}

// Analytics Data

func (s *State) GetAnalytics(userID int64) (analyticsState, bool) {
	s.analyticsMu.RLock()
	defer s.analyticsMu.RUnlock()

	data, ok := s.analyticsStates[userID]
	return data, ok
}

func (s *State) SetAnalytics(userID int64, data analyticsState) {
	s.analyticsMu.Lock()
	defer s.analyticsMu.Unlock()
	s.analyticsStates[userID] = data
}

func (s *State) ClearAnalytics(userID int64) {
	s.analyticsMu.Lock()
	defer s.analyticsMu.Unlock()
	delete(s.analyticsStates, userID)
}

// Edit State

func (s *State) GetEditState(userID int64) (editState, bool) {
	s.editMu.RLock()
	defer s.editMu.RUnlock()

	data, ok := s.editStates[userID]
	return data, ok
}

func (s *State) SetEditState(userID int64, data editState) {
	s.editMu.Lock()
	defer s.editMu.Unlock()
	s.editStates[userID] = data
}

func (s *State) ClearEditState(userID int64) {
	s.editMu.Lock()
	defer s.editMu.Unlock()
	delete(s.editStates, userID)
}
