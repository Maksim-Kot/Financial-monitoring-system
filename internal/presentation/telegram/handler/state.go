package handler

import (
	"sync"
)

type userStep string

const (
	stepIdle             userStep = "idle"
	stepWaitText         userStep = "wait_text"
	stepWaitPhoto        userStep = "wait_photo"
	stepWaitDate         userStep = "wait_date"
	stepWaitOrganisation userStep = "wait_org"
	stepWaitSaveOrg      userStep = "wait_save_org"
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

type State struct {
	userSteps     map[int64]userStep
	listStates    map[int64]listPaginationState
	saveOrgStates map[int64]saveOrgData

	stepsMu       sync.RWMutex
	listMu        sync.RWMutex
	saveOrgMu     sync.RWMutex
	manualItemsMu sync.RWMutex
	analyticsMu   sync.RWMutex
}

func NewState() *State {
	return &State{
		userSteps:     make(map[int64]userStep),
		listStates:    make(map[int64]listPaginationState),
		saveOrgStates: make(map[int64]saveOrgData),
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
