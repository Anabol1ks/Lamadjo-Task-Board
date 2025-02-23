package models

import (
	"gorm.io/gorm"
)

// User представляет зарегистрированного пользователя (руководитель или участник).
type User struct {
	gorm.Model
	TelegramID string `gorm:"uniqueIndex;not null"` // Уникальный идентификатор Telegram
	Name       string `gorm:"not null"`
	Role       string `gorm:"not null"` // "manager" или "member"
	TeamID     *uint  // Для участников — ID команды, к которой они принадлежат
	Teams      []Team `gorm:"foreignKey:ManagerID"` // Для руководителя — список управляемых команд
}
