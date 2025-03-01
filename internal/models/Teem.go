package models

import (
	"time"

	"gorm.io/gorm"
)

// Team представляет команду, созданную руководителем.
type Team struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	ManagerID   uint         `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // ID пользователя-руководителя
	InviteLinks []InviteLink `gorm:"foreignKey:TeamID"`
	Members     []User       `gorm:"foreignKey:TeamID"` // Участники команды
	Tasks       []Task       `gorm:"foreignKey:TeamID"` // Задачи, связанные с командой
	Meetings    []Meeting    `gorm:"foreignKey:TeamID"` // Встречи команды
}

type InviteLink struct {
	ID        uint      `gorm:"primaryKey"`
	Code      string    `gorm:"uniqueIndex;not null"`
	TeamID    uint      `gorm:"not null;index;constraint:OnDelete:CASCADE;"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
