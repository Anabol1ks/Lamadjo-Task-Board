package models

import (
	"time"

	"gorm.io/gorm"
)

// Task представляет задачу, которая может быть назначена команде или конкретному участнику.
type Task struct {
	gorm.Model
	Title       string `gorm:"not null"`
	Description string
	Deadline    time.Time // Срок выполнения
	Status      string    `gorm:"not null; default:'assigned'"` // Статусы: assigned, in_progress, completed
	IsTeam      bool      `gorm:"default:false"`                // true — задача для всей команды
	AssignedTo  *string   // ID пользователя (nil, если IsTeam = true)
	CreatedBy   uint      `gorm:"not null"` // ID создателя
	TeamID      uint      `gorm:"not null"` // ID команды
}
