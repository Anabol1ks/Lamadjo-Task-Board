package models

import "gorm.io/gorm"

// Notification представляет уведомление, отправляемое пользователю.
type Notification struct {
	gorm.Model
	UserID  uint   `gorm:"not null"`      // ID пользователя-получателя
	Message string `gorm:"not null"`      // Текст уведомления
	IsRead  bool   `gorm:"default:false"` // Статус прочтения
	Type    string `gorm:"not null"`      // Тип уведомления, например, "task_update", "meeting_schedule", "invitation"
}
