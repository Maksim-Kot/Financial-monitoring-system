package inmemory

import (
	"context"
	"errors"
	"sync"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/repository"
)

type UserStateRepository struct {
	mu     sync.RWMutex
	states map[int64]*entity.UserState
}

func NewUserStateRepository() *UserStateRepository {
	return &UserStateRepository{
		states: make(map[int64]*entity.UserState),
	}
}

func (r *UserStateRepository) Get(_ context.Context, in repository.UserStateRepositoryGetIn) (repository.UserStateRepositoryGetOut, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	st, ok := r.states[in.UserID]
	if !ok {
		return repository.UserStateRepositoryGetOut{}, nil
	}

	return repository.UserStateRepositoryGetOut{UserState: cloneUserState(st)}, nil
}

func (r *UserStateRepository) Save(_ context.Context, in repository.UserStateRepositorySaveIn) (repository.UserStateRepositorySaveOut, error) {
	if in.UserState == nil {
		return repository.UserStateRepositorySaveOut{}, errors.New("user state is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.states[in.UserState.UserID] = cloneUserState(in.UserState)

	return repository.UserStateRepositorySaveOut{}, nil
}

func (r *UserStateRepository) Delete(_ context.Context, in repository.UserStateRepositoryDeleteIn) (repository.UserStateRepositoryDeleteOut, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.states, in.UserID)

	return repository.UserStateRepositoryDeleteOut{}, nil
}

func cloneUserState(s *entity.UserState) *entity.UserState {
	if s == nil {
		return nil
	}
	out := *s
	if len(s.DraftItems) > 0 {
		out.DraftItems = make([]entity.DraftItem, len(s.DraftItems))
		copy(out.DraftItems, s.DraftItems)
	}
	return &out
}
