package models

import "gorm.io/gorm"

// Team представляет команду, созданную руководителем.
type Team struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	ManagerID   uint      `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // ID пользователя-руководителя
	InviteLink  string    `gorm:"uniqueIndex;not null"`                                   // Уникальная ссылка для приглашения участников
	Members     []User    `gorm:"foreignKey:TeamID"`                                      // Участники команды
	Tasks       []Task    `gorm:"foreignKey:TeamID"`                                      // Задачи, связанные с командой
	Meetings    []Meeting `gorm:"foreignKey:TeamID"`                                      // Встречи команды
}
