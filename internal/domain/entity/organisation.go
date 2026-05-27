package entity

import (
	"fms-project/internal/domain/valueobject"
)

type Organisation struct {
	ID     valueobject.UUID
	UserID int64
	Name   string
}

func NewOrganisation(userID int64, name string) Organisation {
	return Organisation{
		ID:     valueobject.NewRandom(),
		UserID: userID,
		Name:   name,
	}
}
