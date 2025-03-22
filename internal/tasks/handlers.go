package tasks

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/models"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/notification"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/response"
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
// @Failure 403 {object} response.ErrorResponse "Только менеджер может создавать задачу"
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Только менеджер может создавать задачу"})
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

// GetTasksHandlres получает список задач для пользователя
// @Summary Получение списка задач
// @Description Получение списка задач для пользователя
// @Tags tasks
// @Accept json
// @Produce json
// @Param telegram_id query string true "Telegram ID пользователя"
// @Success 200 {object} []response.TaskResponse "Список задач"
// @Failure 400 {object} response.ErrorResponse "telegram_id is required"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} response.ErrorResponse "Ошибка при получении задач"
// @Router /tasks [get]
func GetTasksHandlres(c *gin.Context) {
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

	var tasks []models.Task
	if err := storage.DB.Where("team_id = ?", user.TeamID).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении задач"})
		return
	}
	if err := storage.DB.Where("assigned_to = ?", user.TelegramID).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении задач"})
		return
	}

	var responseTasks []response.TaskResponse
	for _, task := range tasks {
		responseTasks = append(responseTasks, response.TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Deadline:    task.Deadline,
			Status:      task.Status,
			IsTeam:      task.IsTeam,
			AssignedTo:  task.AssignedTo,
			CreatedBy:   task.CreatedBy,
			TeamID:      task.TeamID,
		})
	}

	c.JSON(http.StatusOK, responseTasks)
}

// DeleteTaskHandler удаляет задачу
// @Summary Удаление задачи
// @Description Удаление задачи менеджером команды
// @Tags tasks
// @Accept json
// @Produce json
// @Param telegram_id query string true "Telegram ID менеджера"
// @Param id path string true "ID задачи"
// @Success 200 {object} response.SuccessResponse "Задача успешно удалена"
// @Failure 400 {object} response.ErrorResponse "Error: telegram_id is required CODE: NOT_TG_ID"
// @Failure 400 {object} response.ErrorResponse "Error: task_id is required CODE: NOT_TASK_ID"
// @Failure 400 {object} response.ErrorResponse "Error: У пользователя нет привязанной команды CODE: NOT_TEAM"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 403 {object} response.ErrorResponse "Только менеджер может удалять задачу"
// @Failure 403 {object} response.ErrorResponse "Задачу создали не вы"
// @Failure 500 {object} response.ErrorResponse "Задача не найдена"
// @Router /tasks/{id} [delete]
func DeleteTaskHandler(c *gin.Context) {
	telegramID := c.Query("telegram_id")
	taskID := c.Param("id")
	if telegramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "telegram_id is required", "code": "NOT_TG_ID"})
		return
	}

	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required", "code": "NOT_TASK_ID"})
		return
	}

	var user models.User
	if err := storage.DB.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}
	if user.Role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Только менеджер может удалить задачу"})
		return
	}
	if user.TeamID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "У пользователя нет привязанной команды", "code": "NOT_TEAM"})
		return
	}

	var task models.Task
	if err := storage.DB.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Задача не найдена"})
		return
	}

	if task.CreatedBy != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Задачу создали не вы"})
		return
	}

	var notificationText string
	if task.IsTeam {
		notificationText = fmt.Sprintf(
			"🚀 *Командная задача отменена!*\n\n"+
				"▫️ *Заголовок:* %s\n"+
				"▫️ *Описание:* \n_%s_\n",
			task.Title,
			task.Description,
		)
		var teamUsers []models.User
		if err := storage.DB.Where("team_id = ?", task.TeamID).Find(&teamUsers).Error; err != nil {
			// Логирование ошибки, но можно продолжать отправку уведомлений тому, кого удалось найти
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
			"🚀 *Вашу задача отменили!*\n\n"+
				"▫️ *Заголовок:* %s\n"+
				"▫️ *Описание:* \n_%s_\n"+
				(task.Title),
			(task.Description),
		)
		var assignedUser models.User
		if err := storage.DB.Where("telegram_id = ?", task.AssignedTo).First(&assignedUser).Error; err != nil {
			fmt.Printf("Пользователь не найден: %v\n", err)
			return
		}

		if assignedUser.TelegramID != "" {
			if err := notification.SendTelegramNotification(assignedUser.TelegramID, notificationText); err != nil {
				fmt.Printf("Ошибка отправки уведомления пользователю %s: %v\n", assignedUser.TelegramID, err)
			}
		}
	}

	if err := storage.DB.Delete(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления задачи"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Задача успешно удалена"})
}

type UpdateTaskStatusInput struct {
	Status         string `json:"status" binding:"required"` // Ожидаемые значения: "in_progress" или "completed"
	CompletionText string `json:"completion_text"`           // Отчёт по выполнению (опционально)
	Attachment     string `json:"attachment"`                // Ссылка на файл или описание вложения (опционально)
}

// UpdateTaskStatusHandler обновляет статус задачи.
// Если статус становится "completed", отправляется уведомление менеджеру команды с данными отчёта и информацией о том, кто выполнил задачу.
// UpdateTaskStatusHandler обновляет статус задачи
// @Summary Обновление статуса задачи
// @Description Обновление статуса задачи участником команды или исполнителем персональной задачи
// @Tags tasks
// @Accept json
// @Produce json
// @Param telegram_id query string true "Telegram ID пользователя"
// @Param id path string true "ID задачи"
// @Param task body UpdateTaskStatusInput true "Данные для обновления статуса"
// @Success 200 {object} response.SuccessResponse "Статус задачи успешно обновлен"
// @Failure 400 {object} response.ErrorCodeResponse "Error: telegram_id is required CODE: NOT_TG_ID"
// @Failure 400 {object} response.ErrorCodeResponse "Error: task_id is required CODE: NOT_TASK_ID"
// @Failure 400 {object} response.ErrorCodeResponse "Error: У пользователя нет привязанной команды CODE: NOT_TEAM"
// @Failure 400 {object} response.ErrorResponse "Неверное значение статуса"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 403 {object} response.ErrorResponse "У вас нет прав для изменения статуса этой задачи"
// @Failure 404 {object} response.ErrorResponse "Задача не найдена"
// @Failure 500 {object} response.ErrorResponse "Ошибка при обновлении статуса задачи"
// @Router /tasks/{id}/status [put]
func UpdateTaskStatusHandler(c *gin.Context) {
	telegramID := c.Query("telegram_id")
	if telegramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "telegram_id is required"})
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	// Поиск пользователя по TelegramID
	var user models.User
	if err := storage.DB.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}

	// Поиск задачи по ID
	var task models.Task
	if err := storage.DB.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
		return
	}

	// Проверка прав: если задача персональная, то изменять статус может только назначенный участник.
	if !task.IsTeam {
		if task.AssignedTo == nil || *task.AssignedTo != user.TelegramID {
			c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для изменения статуса этой задачи"})
			return
		}
	} else {
		// Для командной задачи проверяем, что пользователь принадлежит команде.
		if user.TeamID == nil || *user.TeamID != task.TeamID {
			c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для изменения статуса этой задачи"})
			return
		}
	}

	// Считываем данные из запроса
	var input UpdateTaskStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, что статус имеет корректное значение.
	if input.Status != "in_progress" && input.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверное значение статуса. Допустимые значения: in_progress, completed"})
		return
	}

	// Обновляем статус задачи в БД
	task.Status = input.Status
	if err := storage.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении статуса задачи"})
		return
	}

	// Если статус изменён на "completed", отправляем уведомление менеджеру.
	if input.Status == "completed" {
		// Предполагаем, что поле CreatedBy в задаче содержит ID менеджера, создавшего задачу.
		var manager models.User
		if err := storage.DB.First(&manager, task.CreatedBy).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"message": "Статус задачи обновлен, но уведомление менеджеру не отправлено"})
			return
		}

		// Формируем текст уведомления с данными отчёта и информацией, кто выполнил задачу.
		notificationText := fmt.Sprintf(
			"✅ *Задача выполнена!*\n\n▫️ *Заголовок:* %s\n▫️ *Описание:* %s\n▫️ *Статус:* выполнено\n▫️ *Выполнил:* %s\n\n*Отчет участника:*\n%s",
			task.Title,
			task.Description,
			user.Name, // Добавляем имя пользователя, который завершил задачу
			input.CompletionText,
		)
		if input.Attachment != "" {
			notificationText += fmt.Sprintf("\n▫️ *Вложение:* %s", input.Attachment)
		}

		// Отправляем уведомление менеджеру через Telegram (асинхронно).
		if manager.TelegramID != "" {
			go func(chatID string) {
				if err := notification.SendTelegramNotification(chatID, notificationText); err != nil {
					fmt.Printf("Ошибка отправки уведомления менеджеру %s: %v\n", chatID, err)
				}
			}(manager.TelegramID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус задачи успешно обновлен"})
}

// @Summary Получить выданные задачи
// @Description Возвращает список задач, созданных менеджером. Отправляет уведомление в Telegram с списком задач или сообщением об их отсутствии.
// @Tags tasks
// @Accept json
// @Produce json
// @Param telegram_id query string true "Telegram ID менеджера"
// @Success 200 {array} response.TaskResponse "Список выданных задач"
// @Failure 400 {object} response.ErrorResponse "telegram_id is required"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 403 {object} response.ErrorResponse "Доступно только для руководителя"
// @Failure 500 {object} response.ErrorResponse "Ошибка при получении задач"
// @Router /tasks/issued [get]
func IssuedTaskHandler(c *gin.Context) {
	telegramID := c.Query("telegram_id")
	if telegramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "telegram_id is required", "code": "NOT_TG_ID"})
		return
	}

	var user models.User
	if err := storage.DB.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}

	if user.Role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступно только для руководителя"})
		return
	}

	var tasks []models.Task
	if err := storage.DB.Where("created_by = ?", user.ID).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении задач"})
		return
	}

	// Формирование уведомления для Telegram
	// var notificationText string
	// if len(tasks) == 0 {
	// 	notificationText = "ℹ️ *Нет выданных задач*"
	// } else {
	// 	notificationText = "📋 *Список выданных задач:*\n\n"
	// 	for i, task := range tasks {
	// 		notificationText += fmt.Sprintf(
	// 			"%d. *%s*\n▫️ Описание: _%s_\n▫️ Дедлайн: %s\n▫️ Статус: %s\n\n",
	// 			i+1,
	// 			task.Title,
	// 			task.Description,
	// 			notification.FormatDeadline(task.Deadline),
	// 			task.Status,
	// 		)
	// 	}
	// }

	// // Асинхронная отправка уведомления
	// if user.TelegramID != "" {
	// 	go func() {
	// 		if err := notification.SendTelegramNotification(user.TelegramID, notificationText); err != nil {
	// 			fmt.Printf("Ошибка отправки уведомления: %v\n", err)
	// 		}
	// 	}()
	// }

	// Формирование ответа API
	var responseTasks []response.TaskResponse
	for _, task := range tasks {
		responseTasks = append(responseTasks, response.TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Deadline:    task.Deadline,
			Status:      task.Status,
			IsTeam:      task.IsTeam,
			AssignedTo:  task.AssignedTo,
			CreatedBy:   task.CreatedBy,
			TeamID:      task.TeamID,
		})
	}

	c.JSON(http.StatusOK, responseTasks)
}
