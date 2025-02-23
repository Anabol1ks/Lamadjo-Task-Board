package models

import (
	"time"

	"gorm.io/gorm"
)

// Meeting представляет встречу, назначенную для команды.
type Meeting struct {
	gorm.Model
	Title          string    `gorm:"not null"`
	MeetingType    string    `gorm:"not null"` // "online" или "offline"
	Date           time.Time `gorm:"not null"` // Дата встречи
	StartTime      time.Time `gorm:"not null"` // Время начала встречи
	EndTime        time.Time `gorm:"not null"` // Время окончания встречи
	ConferenceLink string    // Для онлайн встреч – ссылка на конференцию
	Room           string    // Для оффлайн встреч – номер/название аудитории
	TeamID         uint      `gorm:"not null"` // ID команды, для которой назначена встреча
	CreatedBy      uint      // ID руководителя, создавшего встречу
}
