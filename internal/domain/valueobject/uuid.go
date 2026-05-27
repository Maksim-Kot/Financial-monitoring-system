package valueobject

import (
	"fmt"

	"github.com/google/uuid"
)

type UUID struct {
	value uuid.UUID
}

func NewUUID(raw any) (UUID, error) {
	v, ok := raw.(string)

	if !ok {
		return UUID{}, fmt.Errorf("invalid value <%v>", raw)
	}

	if v == "" {
		return UUID{}, fmt.Errorf("invalid value <%v>", v)
	}

	uuid, err := uuid.Parse(v)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid value <%v>", v)
	}

	return UUID{value: uuid}, nil
}

func NewRandom() UUID {
	return UUID{value: uuid.New()}
}

func (u UUID) String() string {
	return u.value.String()
}

func (u UUID) IsNil() bool {
	return u.value == uuid.Nil
}

func (u UUID) Equal(another UUID) bool {
	return u.String() == another.String()
}

func (u UUID) Value() uuid.UUID {
	return u.value
}
