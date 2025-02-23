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
	Deadline    time.Time // Срок выполнения задачи
	Status      string    `gorm:"not null"` // Например, "assigned", "in_progress", "completed"
	AssignedTo  uint      // ID пользователя, если задача персональная
	CreatedBy   uint      // ID руководителя, который создал задачу
	TeamID      uint      `gorm:"not null"` // ID команды, к которой относится задача
}
