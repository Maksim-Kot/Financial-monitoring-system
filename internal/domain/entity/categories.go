package entity

import (
	"fms-project/internal/domain/valueobject"
)

type Category struct {
	ID   valueobject.UUID
	Name string
	Icon string
}

func NewCategory(name string, icon string) Category {
	return Category{
		ID:   valueobject.NewRandom(),
		Name: name,
		Icon: icon,
	}
}

type categorySeed struct {
	Name string
	Icon string
}

var DefaultCategorySeeds = []categorySeed{
	{Name: "Продукты", Icon: "🍎"},
	{Name: "Гигиена и уход", Icon: "🧼"},
	{Name: "Товары для дома", Icon: "🏠"},
	{Name: "Бытовая химия", Icon: "🧴"},
	{Name: "Одежда и обувь", Icon: "👕"},
	{Name: "Техника и электроника", Icon: "💻"},
	{Name: "Здоровье", Icon: "💊"},
	{Name: "Развлечения", Icon: "🎥"},
	{Name: "Другое", Icon: "💰"},
}
