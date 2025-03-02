package meetings

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

// TimeSlot определяет фиксированный временной интервал.
type TimeSlot struct {
	Start string `json:"start"` // Например, "12:00"
	End   string `json:"end"`   // Например, "13:20"
}

// FixedTimeSlots – список фиксированных временных блоков для офлайн встреч.
var FixedTimeSlots = []TimeSlot{
	{"12:00", "13:20"},
	{"13:30", "14:50"},
	{"15:00", "16:20"},
	{"16:30", "17:50"},
}

type CreateMeetingInput struct {
	Title       string `json:"title" binding:"required"`
	MeetingType string `json:"meeting_type" binding:"required"` // "online" или "offline"
	Date        string `json:"date" binding:"required"`         // Формат "YYYY-MM-DD"
	StartTime   string `json:"start_time" binding:"required"`   // Формат "HH:MM"
	EndTime     string `json:"end_time" binding:"required"`     // Формат "HH:MM"
	Room        string `json:"room"`                            // Обязательное для офлайн встреч
}

// CreateMeetingHandler создаёт встречу
// @Summary Создание встречи
// @Description Создает новую встречу для команды. Доступно только для менеджеров.
// @Tags meetings
// @Accept json
// @Produce json
// @Param telegram_id query string true "Уникальный идентификатор Telegram"
// @Param input body CreateMeetingInput true "Данные встречи"
// @Success 200 {object} response.MeetingResponse "Информация о созданной встрече"
// @Failure 400 {object} response.ErrorResponse "Ошибка валидации или некорректные данные"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 403 {object} response.ErrorResponse "Доступ запрещен (не менеджер)"
// @Failure 409 {object} response.ErrorResponse "Конфликт по времени и аудитории"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /meetings [post]
func CreateMeetingHandler(c *gin.Context) {
	telegramID := c.Query("telegram_id")
	if telegramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "telegram_id is required"})
		return
	}

	// Поиск пользователя по TelegramID
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

	var input CreateMeetingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Парсинг даты и времени
	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты (YYYY-MM-DD)"})
		return
	}
	parsedStart, err := time.Parse("15:04", input.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат времени (HH:MM)"})
		return
	}
	parsedEnd, err := time.Parse("15:04", input.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат времени (HH:MM)"})
		return
	}

	// Формируем полные временные метки
	startDateTime := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
		parsedStart.Hour(), parsedStart.Minute(), 0, 0, time.Local)
	endDateTime := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
		parsedEnd.Hour(), parsedEnd.Minute(), 0, 0, time.Local)
	if endDateTime.Before(startDateTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Время окончания не может быть раньше времени начала"})
		return
	}

	// Если встреча офлайн – проверяем, что время соответствует фиксированным слотам и что аудитория существует
	if input.MeetingType == "offline" {
		if input.Room == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Для офлайн встречи необходимо указать аудиторию (room)"})
			return
		}
		// Проверяем, существует ли указанная аудитория в БД
		var room models.Room
		if err := storage.DB.Where("name = ?", input.Room).First(&room).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Аудитория не найдена"})
			return
		}
		// Проверка, соответствует ли заданное время одному из фиксированных слотов
		slotMatched := false
		for _, slot := range FixedTimeSlots {
			slotStart, err := time.Parse("15:04", slot.Start)
			if err != nil {
				continue
			}
			slotEnd, err := time.Parse("15:04", slot.End)
			if err != nil {
				continue
			}
			tsStart := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
				slotStart.Hour(), slotStart.Minute(), 0, 0, time.Local)
			tsEnd := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
				slotEnd.Hour(), slotEnd.Minute(), 0, 0, time.Local)
			if startDateTime.Equal(tsStart) && endDateTime.Equal(tsEnd) {
				slotMatched = true
				break
			}
		}
		if !slotMatched {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Время встречи должно соответствовать одному из фиксированных временных блоков"})
			return
		}

		// Проверка конфликтов: ищем встречи в той же аудитории, в ту же дату и с пересекающимся интервалом.
		var existingMeetings []models.Meeting
		if err := storage.DB.Where("team_id = ? AND meeting_type = ? AND date = ? AND room = ? AND ((start_time < ? AND end_time > ?))",
			*user.TeamID, "offline", parsedDate, input.Room, endDateTime, startDateTime).Find(&existingMeetings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки конфликтов"})
			return
		}
		if len(existingMeetings) > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Конфликт по времени и аудитории"})
			return
		}
	}

	// Для онлайн встреч можно сгенерировать ссылку (пример)
	var confLink string
	if input.MeetingType == "online" {
		confLink = "https://zoom.us/j/ТИПО_ССЫЛКА_НА_ЗУМ"
	}

	meeting := models.Meeting{
		Title:          input.Title,
		MeetingType:    input.MeetingType,
		Date:           parsedDate,
		StartTime:      startDateTime,
		EndTime:        endDateTime,
		ConferenceLink: confLink,
		Room:           input.Room,
		TeamID:         *user.TeamID,
		CreatedBy:      user.ID,
	}

	if err := storage.DB.Create(&meeting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании встречи"})
		return
	}

	response := response.MeetingResponse{
		ID:             meeting.ID,
		Title:          meeting.Title,
		MeetingType:    meeting.MeetingType,
		Date:           meeting.Date,
		StartTime:      meeting.StartTime,
		EndTime:        meeting.EndTime,
		ConferenceLink: meeting.ConferenceLink,
		Room:           meeting.Room,
		TeamID:         meeting.TeamID,
		CreatedBy:      meeting.CreatedBy,
		CreatedAt:      meeting.CreatedAt,
		UpdatedAt:      meeting.UpdatedAt,
	}

	var notificationText string
	if meeting.MeetingType == "online" {
		notificationText = fmt.Sprintf(
			"📢 *Новая встреча!*\n\n"+
				"*Название:* %s\n"+
				"*Дата:* %s\n"+
				"*Время:* %s - %s\n"+
				"*Тип:* Онлайн\n"+
				"*Ссылка:* [Подключиться](%s)",
			meeting.Title,
			formatDateRussian(meeting.Date),
			meeting.StartTime.Format("15:04"),
			meeting.EndTime.Format("15:04"),
			meeting.ConferenceLink,
		)
	} else {
		notificationText = fmt.Sprintf(
			"📢 *Новая встреча!*\n\n"+
				"*Название:* %s\n"+
				"*Дата:* %s\n"+
				"*Время:* %s - %s\n"+
				"*Тип:* Офлайн\n"+
				"*Аудитория:* %s",
			meeting.Title,
			formatDateRussian(meeting.Date),
			meeting.StartTime.Format("15:04"),
			meeting.EndTime.Format("15:04"),
			meeting.Room,
		)
	}

	// Получаем всех участников команды (предполагается, что в модели User есть поле TelegramID и TeamID)
	var teamUsers []models.User
	if err := storage.DB.Where("team_id = ?", meeting.TeamID).Find(&teamUsers).Error; err != nil {
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

	c.JSON(http.StatusOK, response)
}

// GetAvailableTimeSlotsHandler получает доступные временные слоты
// @Summary Получение доступных временных слотов
// @Description Возвращает список доступных временных слотов для указанной аудитории на выбранную дату
// @Tags meetings
// @Accept json
// @Produce json
// @Param room query string true "Номер аудитории"
// @Param date query string true "Дата в формате YYYY-MM-DD"
// @Success 200 {array} []TimeSlot "Список доступных временных слотов"
// @Failure 400 {object} response.ErrorResponse "Отсутствуют обязательные параметры или неверный формат"
// @Failure 500 {object} response.ErrorResponse "Ошибка при получении данных"
// @Router /meetings/available-slots [get]
func GetAvailableTimeSlotsHandler(c *gin.Context) {
	roomName := c.Query("room")
	dateStr := c.Query("date")
	if roomName == "" || dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room и date обязательны"})
		return
	}
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты (YYYY-MM-DD)"})
		return
	}

	// Проверяем, существует ли аудитория
	var room models.Room
	if err := storage.DB.Where("name = ?", roomName).First(&room).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Аудитория не найдена"})
		return
	}

	// Получаем все встречи для данной аудитории и даты.
	var meetings []models.Meeting
	if err := storage.DB.Where("room = ? AND date = ?", roomName, parsedDate).Find(&meetings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка запроса к БД"})
		return
	}

	// Собираем занятые интервалы.
	occupiedSlots := []struct {
		Start time.Time
		End   time.Time
	}{}
	for _, m := range meetings {
		occupiedSlots = append(occupiedSlots, struct {
			Start time.Time
			End   time.Time
		}{Start: m.StartTime, End: m.EndTime})
	}

	// Определяем доступные фиксированные слоты.
	availableSlots := []TimeSlot{}
	for _, slot := range FixedTimeSlots {
		slotStartParsed, err := time.Parse("15:04", slot.Start)
		if err != nil {
			continue
		}
		slotEndParsed, err := time.Parse("15:04", slot.End)
		if err != nil {
			continue
		}
		slotStart := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
			slotStartParsed.Hour(), slotStartParsed.Minute(), 0, 0, time.Local)
		slotEnd := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
			slotEndParsed.Hour(), slotEndParsed.Minute(), 0, 0, time.Local)

		conflict := false
		for _, occ := range occupiedSlots {
			if slotStart.Before(occ.End) && slotEnd.After(occ.Start) {
				conflict = true
				break
			}
		}
		if !conflict {
			availableSlots = append(availableSlots, slot)
		}
	}

	c.JSON(http.StatusOK, gin.H{"available_slots": availableSlots})
}

// DeleteMeetingHandler удаляет встречу
// @Summary Удаление встречи
// @Description Удаляет встречу из расписания команды. Доступно только для менеджеров.
// @Tags meetings
// @Accept json
// @Produce json
// @Param telegram_id query string true "Уникальный идентификатор Telegram"
// @Param id path string true "ID встречи"
// @Success 200 {object} response.SuccessResponse "Встреча успешно удалена"
// @Failure 400 {object} response.ErrorResponse "Отсутствует telegram_id или ID встречи"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 403 {object} response.ErrorResponse "Доступ запрещен (не менеджер)"
// @Failure 404 {object} response.ErrorResponse "Встреча не найдена"
// @Failure 500 {object} response.ErrorResponse "Ошибка при удалении встречи"
// @Router /meetings/{id} [delete]
func DeleteMeetingHandler(c *gin.Context) {
	telegramID := c.Query("telegram_id")
	if telegramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "telegram_id is required"})
		return
	}

	meetingID := c.Param("id")
	if meetingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID встречи не указан"})
		return
	}

	var user models.User
	if err := storage.DB.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}

	if user.Role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Только менеджер может удалять встречи"})
		return
	}

	var meeting models.Meeting
	if err := storage.DB.First(&meeting, meetingID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Встреча не найдена"})
		return
	}

	if user.TeamID == nil || *user.TeamID != meeting.TeamID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к данной встрече"})
		return
	}

	meetingTitle := meeting.Title

	if err := storage.DB.Delete(&meeting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении встречи"})
		return
	}

	// Формируем текст уведомления
	notificationText := fmt.Sprintf(
		"❌ *Встреча отменена!*\n\n"+
			"*Название:* %s\n"+
			"*Дата:* %s\n"+
			"*Время:* %s - %s",
		meetingTitle,
		formatDateRussian(meeting.Date),
		meeting.StartTime.Format("15:04"),
		meeting.EndTime.Format("15:04"),
	)

	// Получаем всех участников команды
	var teamUsers []models.User
	if err := storage.DB.Where("team_id = ?", meeting.TeamID).Find(&teamUsers).Error; err != nil {
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

	c.JSON(http.StatusOK, gin.H{"message": "Встреча успешно удалена"})
}

// GetMyMeeting получает список встреч команды пользователя
// @Summary Получение встреч команды
// @Description Возвращает список всех встреч команды, к которой привязан пользователь
// @Tags meetings
// @Accept json
// @Produce json
// @Param telegram_id query string true "Уникальный идентификатор Telegram"
// @Success 200 {array} response.MeetingResponse "Список встреч команды"
// @Failure 400 {object} response.ErrorResponse "Отсутствует telegram_id или у пользователя нет команды"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} response.ErrorResponse "Ошибка при получении встреч"
// @Router /meetings/my [get]
func GetMyMeeting(c *gin.Context) {
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

	if user.TeamID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "У пользователя нет привязанной команды"})
		return
	}

	var meetings []models.Meeting
	if err := storage.DB.Where("team_id = ?", *user.TeamID).Find(&meetings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении встреч"})
		return
	}

	c.JSON(http.StatusOK, meetings)
}

func formatDateRussian(t time.Time) string {
	months := []string{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}
	day := t.Day()
	month := months[t.Month()-1] // t.Month() возвращает значение от 1 до 12
	year := t.Year()
	return fmt.Sprintf("%d %s %d", day, month, year)
}
