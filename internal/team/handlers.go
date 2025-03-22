package team

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/models"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/notification"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateTeamInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func generateInviteLink() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 16
	invite := make([]byte, length)
	for i := 0; i < length; i++ {
		// Получаем случайное число от 0 до len(charset)-1
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		invite[i] = charset[num.Int64()]
	}
	return string(invite), nil
}

// CreateTeamHandler создаёт команду. Создавать команду может только менеджер.
// @Summary Создание команды
// @Description Создает команду, если запрос исходит от пользователя с ролью manager.
// @Tags team
// @Accept json
// @Produce json
// @Param telegram_id query string true "Уникальный идентификатор Telegram"
// @Param input body CreateTeamInput true "Данные команды"
// @Success 200 {object} response.TeamResponse "Информация о созданной команде"
// @Failure 400 {object} response.ErrorResponse "Ошибка валидации или отсутствует telegram_id"
// @Failure 403 {object} response.ErrorResponse "Доступ запрещен (не менеджер)"
// @Failure 500 {object} response.ErrorResponse "Ошибка создания команды"
// @Router /team [post]
func CreateTeamHandler(c *gin.Context) {
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Только менеджер может создать команду"})
		return
	}

	var input CreateTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team := models.Team{
		Name:        input.Name,
		Description: input.Description,
		ManagerID:   user.ID,
	}

	if err := storage.DB.Create(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании команды"})
		return
	}

	user.TeamID = &team.ID
	if err := storage.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении пользователя"})
		return
	}

	c.JSON(http.StatusOK, team)
}

type InviteJoinRequest struct {
	TelegramID string `json:"telegram_id" binding:"required"`
	InviteCode string `json:"invite_code" binding:"required"`
}

// JoinTeamHandler позволяет пользователю присоединиться к команде, используя пригласительный код.
// @Summary Присоединение к команде
// @Description Позволяет пользователю присоединиться к команде, используя пригласительный код.
// @Tags team
// @Accept json
// @Produce json
// @Param input body InviteJoinRequest true "Данные для присоединения к команде"
// @Success 200 {object} response.SuccessResponse "Успешное присоединение к команде"
// @Failure 400 {object} response.ErrorResponse "Ошибка валидации"
// @Failure 404 {object} response.ErrorCodeResponse "Error: Команда не найдена CODE: INVITE_CODE_INVALID, Error: Команда не найдена. CODE: TEAM_NOT_FOUND, Error: Пользователь не найден. Зарегистрируйтесь через бота. CODE: USER_NOT_FOUND"
// @Failure 409 {object} response.ErrorResponse "Вы уже присоединились к этой команде"
// @Failure 500 {object} response.ErrorResponse "Ошибка при присоединении к команде"
// @Router /team/join [post]
func JoinTeamHandler(c *gin.Context) {
	var req InviteJoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ищем активную ссылку
	var invite models.InviteLink
	if err := storage.DB.Where("code = ? AND expires_at > ?", req.InviteCode, time.Now()).First(&invite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Неверный или просроченный код приглашения",
			"code":  "INVITE_CODE_INVALID",
		})
		return
	}

	var team models.Team
	if err := storage.DB.First(&team, invite.TeamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Команда не найдена", "code": "TEAM_NOT_FOUND"})
		return
	}

	var user models.User
	if err := storage.DB.Where("telegram_id = ?", req.TelegramID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден. Зарегистрируйтесь через бота.", "code": "USER_NOT_FOUND"})
		return
	}

	if user.TeamID != nil && *user.TeamID == team.ID {
		c.JSON(http.StatusConflict, gin.H{"message": "Вы уже присоединились к этой команде"})
		return
	}

	user.TeamID = &team.ID
	if err := storage.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при присоединении к команде"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Вы успешно присоединились к команде", "team": team})
}

// GetLinkTeamHandler получает ссылку-приглашение для команды
// @Summary Получение ссылки-приглашения
// @Description Возвращает ссылку для приглашения новых участников в команду. Доступно только для менеджеров.
// @Tags team
// @Accept json
// @Produce json
// @Param telegram_id query string true "Уникальный идентификатор Telegram"
// @Success 200 {string} string "URL ссылки-приглашения"
// @Failure 400 {object} response.ErrorResponse "Отсутствует telegram_id"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 403 {object} response.ErrorResponse "Доступ запрещен (не менеджер)"
// @Failure 404 {object} response.ErrorCodeResponse "Error:Отсутствует команда у пользователя Code:USER_HAS_NO_TEAM, Error:Команда не найдена Code:TEAM_NOT_FOUND"
// @Failure 500 {object} response.ErrorResponse "Ошибка при создании/получении ссылки-приглашения"
// @Router /team/invite [get]
func GetLinkTeamHandler(c *gin.Context) {
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Только менеджер может получить ссылку на приглашение в команду"})
		return
	}

	if user.TeamID == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Отсутствует команда у пользователя",
			"code":  "USER_HAS_NO_TEAM",
		})
		return
	}

	var team models.Team
	if err := storage.DB.Where("id = ?", user.TeamID).First(&team).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Команда не найдена",
			"code":  "TEAM_NOT_FOUND",
		})
		return
	}

	code, err := generateInviteLink()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации ссылки"})
		return
	}

	// Сохраняем ссылку с временем жизни
	expiresAt := time.Now().Add(24 * time.Hour)
	invite := models.InviteLink{
		Code:      code,
		TeamID:    team.ID,
		ExpiresAt: expiresAt,
	}

	if err := storage.DB.Create(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения ссылки"})
		return
	}

	// Формируем URL
	urlLink := fmt.Sprintf("http://t.me/LamadjoTask_bot?start=%s", code)
	c.JSON(http.StatusOK, urlLink)
}

// GetMyTeamHandler получает информацию о команде текущего пользователя
// @Summary Получение информации о своей команде
// @Description Возвращает данные о команде, к которой принадлежит пользователь
// @Tags team
// @Accept json
// @Produce json
// @Param telegram_id query string true "Уникальный идентификатор Telegram"
// @Success 200 {object} response.TeamResponse "Информация о команде"
// @Failure 400 {object} response.ErrorResponse "Отсутствует telegram_id"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 404 {object} response.ErrorCodeResponse "Error:Отсутствует команда у пользователя Сode:USER_HAS_NO_TEAM, Error: Команда не найдена Сode:TEAM_NOT_FOUND"
// @Router /team/my [get]
func GetMyTeamHandler(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Отсутствует команда у пользователя",
			"code":  "USER_HAS_NO_TEAM",
		})
		return
	}

	var team models.Team
	if err := storage.DB.Where("id = ?", user.TeamID).First(&team).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Команда не найдена",
			"code":  "TEAM_NOT_FOUND",
		})
	}
	c.JSON(http.StatusOK, team)
}

// ChangeTeamHandler изменяет информацию о команде
// @Summary Изменение информации о команде
// @Description Обновляет название и описание команды. Доступно только для менеджеров.
// @Tags team
// @Accept json
// @Produce json
// @Param telegram_id query string true "Уникальный идентификатор Telegram"
// @Param input body CreateTeamInput true "Данные для обновления команды"
// @Success 200 {object} response.SuccessResponse "Команда успешно обновлена"
// @Failure 400 {object} response.ErrorResponse "Ошибка валидации или отсутствует telegram_id"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 403 {object} response.ErrorResponse "Доступ запрещен (не менеджер)"
// @Failure 404 {object} response.ErrorCodeResponse "Error:Отсутствует команда у пользователя Code:USER_HAS_NO_TEAM, Error:Команда не найдена Code:TEAM_NOT_FOUND"
// @Failure 500 {object} response.ErrorResponse "Ошибка при обновлении команды"
// @Router /team [put]
func ChangeTeamHandler(c *gin.Context) {
	telegramID := c.Query("telegram_id")
	if telegramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "telegram_id is required"})
		return
	}

	var input CreateTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := storage.DB.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}

	if user.Role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Только менеджер может получить ссылку на приглашение в команду"})
		return
	}

	if user.TeamID == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Отсутствует команда у пользователя",
			"code":  "USER_HAS_NO_TEAM",
		})
		return
	}

	var team models.Team
	if err := storage.DB.Where("id = ?", user.TeamID).First(&team).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Команда не найдена",
			"code":  "TEAM_NOT_FOUND",
		})
	}

	team.Name = input.Name
	team.Description = input.Description

	if err := storage.DB.Save(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении команды"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Команда успешно обновлена"})
}

// GetMembersTeam получает список участников команды
// @Summary Получение списка участников команды
// @Description Возвращает список всех участников команды, кроме текущего пользователя. Доступно только для менеджеров.
// @Tags team
// @Accept json
// @Produce json
// @Param telegram_id query string true "Уникальный идентификатор Telegram"
// @Success 200 {array} response.UserResponse "Список участников команды"
// @Failure 400 {object} response.ErrorResponse "Отсутствует telegram_id или ошибка валидации"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 403 {object} response.ErrorResponse "Доступ запрещен (не менеджер)"
// @Failure 404 {object} response.ErrorCodeResponse "Error:Отсутствует команда у пользователя Code:USER_HAS_NO_TEAM, Error:Команда не найдена Code:TEAM_NOT_FOUND"
// @Failure 500 {object} response.ErrorResponse "Ошибка при получении участников команды"
// @Router /team/members [get]
func GetMembersTeam(c *gin.Context) {
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Только менеджер может получить ссылку на приглашение в команду"})
		return
	}

	if user.TeamID == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Отсутствует команда у пользователя",
			"code":  "USER_HAS_NO_TEAM",
		})
		return
	}

	var team models.Team
	if err := storage.DB.Where("id = ?", user.TeamID).First(&team).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Команда не найдена",
			"code":  "TEAM_NOT_FOUND",
		})
	}

	var member []models.User
	if err := storage.DB.Where("team_id = ? AND telegram_id != ?", user.TeamID, telegramID).Find(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении участников команды"})
		return
	}
	c.JSON(http.StatusOK, member)
}

// LeaveMemberTeamHandler позволяет пользователю покинуть текущую команду
// @Summary Покинуть команду
// @Description Позволяет пользователю выйти из текущей команды
// @Tags team
// @Accept json
// @Produce json
// @Param telegram_id query string true "Уникальный идентификатор Telegram"
// @Success 200 {object} response.SuccessResponse "Команда успешно покинута"
// @Failure 400 {object} response.ErrorResponse "Отсутствует telegram_id"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 403 {object} response.ErrorResponse "Manager не может просто так покинуть команду"
// @Failure 404 {object} response.ErrorCodeResponse "Error: Отсутствует команда у пользователя Code: USER_HAS_NO_TEAM"
// @Failure 500 {object} response.ErrorResponse "Ошибка при попытке покинуть команду"
// @Router /team/leave [get]
func LeaveMemberTeamHandler(c *gin.Context) {
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

	if user.Role == "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Manager не может просто так покинуть команду"})
		return
	}

	if user.TeamID == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Отсутствует команда у пользователя",
			"code":  "USER_HAS_NO_TEAM",
		})
		return
	}

	user.TeamID = nil
	if err := storage.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при попытке покинуть команду"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Команда покинута"})
}

// KickMemberTeamHandler позволяет менеджеру исключить участника из команды
// @Summary Исключить участника из команды
// @Description Позволяет менеджеру исключить участника из команды
// @Tags team
// @Accept json
// @Produce json
// @Param telegram_id query string true "Уникальный идентификатор Telegram"
// @Param kick_telegram_id query string true "Уникальный идентификатор Telegram участника, который будет исключен"
// @Success 200 {object} response.SuccessResponse "Участник успешно исключен из команды"
// @Failure 400 {object} response.ErrorResponse "Отсутствует telegram_id или kick_telegram_id"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 403 {object} response.ErrorResponse "Error: Только менеджер может исключить участника из команды. CODE: NOT_MANAGER, Error: Пользователь не находится в вашей команде, CODE: NOT_IN_TEAM"
// @Failure 500 {{object} response.ErrorResponse "Ошибка при попытке исключить участника из команды"
// @Router /team/kick [get]
func KickMemberTeamHandler(c *gin.Context) {
	telegramID := c.Query("telegram_id")
	if telegramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "telegram_id is required"})
		return
	}
	kickTelegramID := c.Query("kick_telegram_id")
	if kickTelegramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kick_telegram_id is required"})
		return
	}
	var user models.User
	if err := storage.DB.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}
	if user.Role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Только менеджер может исключить участника из команды", "code": "NOT_MANAGER"})
		return
	}
	var userKick models.User
	if err := storage.DB.Where("telegram_id = ?", kickTelegramID).First(&userKick).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}

	if *userKick.TeamID != *user.TeamID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Пользователь не находится в вашей команде", "code": "NOT_IN_TEAM"})
		return
	}

	userKick.TeamID = nil

	if err := storage.DB.Save(&userKick).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при попытке исключить участника из команды"})
		return
	}

	notificationText := fmt.Sprintf(
		"😕 *Мы сожелеем, но... *\n\n" +
			"Вы были исключены из команды...",
	)
	if err := notification.SendTelegramNotification(userKick.TelegramID, notificationText); err != nil {
		fmt.Printf("Ошибка отправки уведомления пользователю %s: %v\n", userKick.TelegramID, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Участник успешно исключен из команды"})
}

// DeleteTeamHandler удаляет команду и очищает связи с участниками
// @Summary Удаление команды
// @Description Удаляет команду и очищает связи со всеми участниками. Доступно только для владельца команды.
// @Tags team
// @Accept json
// @Produce json
// @Param telegram_id query string true "Уникальный идентификатор Telegram"
// @Success 200 {object} response.SuccessResponse "Команда успешно удалена"
// @Failure 400 {object} response.ErrorResponse "Отсутствует telegram_id"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 403 {object} response.ErrorCodeResponse "Error: Только руководитель может удалить команду Code: ONLY_MANAGER_DELETE_TEAM, Error: Вы не являетесь владельцем команды Code: NOT_OWNER_OF_TEAM"
// @Failure 500 {object} response.ErrorResponse "Ошибка при удалении команды"
// @Router /team [delete]
func DeleteTeamHandler(c *gin.Context) {
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Только руководитель может удалить команду", "code": "ONLY_MANAGER_DELETE_TEAM"})
		return
	}

	var team models.Team
	if err := storage.DB.Where("manager_id = ?", user.ID).First(&team).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Вы не являетесь владельцем команды", "code": "NOT_OWNER_OF_TEAM"})
		return
	}

	// Удаляем команду и обнуляем TeamID у участников
	err := storage.DB.Transaction(func(tx *gorm.DB) error {
		// Обнуляем TeamID у всех участников команды
		if err := tx.Model(&models.User{}).
			Where("team_id = ?", team.ID).
			Update("team_id", nil).
			Error; err != nil {
			return err
		}

		// Удаляем саму команду
		if err := tx.Delete(&team).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении команды"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Команда успешно удалена"})
}
