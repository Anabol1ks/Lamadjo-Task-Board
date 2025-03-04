package tasks

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/models"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/notification"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/storage"
	"github.com/gin-gonic/gin"
)

type TaskInput struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description" binding:"required"`
	Deadline    time.Time `json:"deadline" binding:"required"` //RFC 3339
	IsTeam      bool      `json:"is_team"`
	AssignedTo  *string   `json:"assigned_to"`
}

// CreateTaskHandlres создает новую задачу
// @Summary Создание задачи
// @Description Создание задачи для команды и индивидуально
// @Tags tasks
// @Accept json
// @Produce json
// @Param telegram_id query string true "Telegram ID управляющегоr"
// @Param task body TaskInput true "Информация задачи"
// @Success 200 {object} response.SuccessResponse "Задача успешно создана"
// @Failure 400 {object} response.ErrorResponse "telegram_id is required"
// @Failure 400 {object} response.ErrorResponse "У пользователя нет привязанной команды"
// @Failure 400 {object} response.ErrorResponse "assigned_to обязателен для персональных задач"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 403 {object} response.ErrorResponse "Только менеджер может создавать встречи"
// @Failure 500 {object} response.ErrorResponse "Ошибка при создании задачи"
// @Router /tasks [post]
func CreateTaskHandlres(c *gin.Context) {
	telegramID := c.Query("telegram_id")
	if telegramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "telegram_id is required"})
		return
	}

	var user models.User
	if err := storage.DB.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}
	if user.Role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Только менеджер может создавать встречи"})
		return
	}
	if user.TeamID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "У пользователя нет привязанной команды"})
		return
	}

	var input TaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !input.IsTeam && input.AssignedTo == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assigned_to обязателен для персональных задач"})
		return
	}

	var task = models.Task{
		Title:       input.Title,
		Description: input.Description,
		Deadline:    input.Deadline,
		IsTeam:      input.IsTeam,
		AssignedTo:  input.AssignedTo,
		TeamID:      *user.TeamID,
		CreatedBy:   user.ID,
	}

	if err := storage.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании задачи"})
		return
	}

	var notificationText string
	if input.IsTeam {
		notificationText = fmt.Sprintf(
			"🚀 *Новая командная задача!*\n\n"+
				"▫️ *Заголовок:* %s\n"+
				"▫️ *Описание:* \n_%s_\n"+
				"▫️ *Дедлайн:* %s\n"+
				"▫️ *Тип:* Общая задача команды\n\n"+
				"🕑 Создано: %s",
			(task.Title),
			(task.Description),
			notification.FormatDeadline(task.Deadline),
			time.Now().Format("02.01.2006 15:04"),
		)

		var teamUsers []models.User
		if err := storage.DB.Where("team_id = ?", user.TeamID).Find(&teamUsers).Error; err != nil {
			fmt.Printf("Ошибка получения участников команды: %v\n", err)
		}

		for _, u := range teamUsers {
			if u.TelegramID != "" {
				go func(chatID string) {
					if err := notification.SendTelegramNotification(chatID, notificationText); err != nil {
						fmt.Printf("Ошибка отправки уведомления пользователю %s: %v\n", chatID, err)
					}
				}(u.TelegramID)
			}
		}
	} else {
		notificationText = fmt.Sprintf(
			"📌 *Новая персональная задача!*\n\n"+
				"▫️ *Заголовок:* %s\n"+
				"▫️ *Описание:* \n_%s_\n"+
				"▫️ *Дедлайн:* %s\n"+
				"▫️ *Назначена:* Вам лично\n\n"+
				"🕑 Создано: %s",
			task.Title,
			task.Description,
			notification.FormatDeadline(task.Deadline),
			time.Now().Format("02.01.2006 15:04"),
		)

		var assignedUser models.User
		if err := storage.DB.Where("telegram_id = ?", input.AssignedTo).First(&assignedUser).Error; err != nil {
			fmt.Printf("Пользователь не найден: %v\n", err)
			return
		}

		if assignedUser.TelegramID != "" {
			if err := notification.SendTelegramNotification(assignedUser.TelegramID, notificationText); err != nil {
				fmt.Printf("Ошибка отправки уведомления пользователю %s: %v\n", assignedUser.TelegramID, err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Задача успешно создана"})
}
