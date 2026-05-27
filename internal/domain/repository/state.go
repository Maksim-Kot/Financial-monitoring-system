package repository

import (
	"context"

	"fms-project/internal/domain/entity"
)

type UserStateRepositoryGetIn struct {
	UserID int64
}

type UserStateRepositoryGetOut struct {
	UserState *entity.UserState
}

func (o *UserStateRepositoryGetOut) Exists() bool {
	return o.UserState != nil
}

type UserStateRepositorySaveIn struct {
	UserState *entity.UserState
}

type UserStateRepositorySaveOut struct{}

type UserStateRepositoryDeleteIn struct {
	UserID int64
}

type UserStateRepositoryDeleteOut struct{}

type UserStateRepository interface {
	Get(ctx context.Context, in UserStateRepositoryGetIn) (UserStateRepositoryGetOut, error)
	Save(ctx context.Context, in UserStateRepositorySaveIn) (UserStateRepositorySaveOut, error)
	Delete(ctx context.Context, in UserStateRepositoryDeleteIn) (UserStateRepositoryDeleteOut, error)
}
