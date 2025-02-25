package team

import (
	"crypto/rand"
	"math/big"
	"net/http"

	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/models"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/storage"
	"github.com/gin-gonic/gin"
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

	inviteLink, err := generateInviteLink()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации ссылки приглашения"})
		return
	}

	team := models.Team{
		Name:        input.Name,
		Description: input.Description,
		ManagerID:   user.ID,
		InviteLink:  inviteLink,
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
// @Failure 404 {object} response.ErrorResponse "Неверный код приглашения или пользователь не найден"
// @Failure 409 {object} response.ErrorResponse "Вы уже присоединились к этой команде"
// @Failure 500 {object} response.ErrorResponse "Ошибка при присоединении к команде"
// @Router /team/join [post]
func JoinTeamHandler(c *gin.Context) {
	var req InviteJoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var team models.Team
	if err := storage.DB.Where("invite_link = ?", req.InviteCode).First(&team).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Неверный код приглашения"})
		return
	}

	var user models.User
	if err := storage.DB.Where("telegram_id = ?", req.TelegramID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден. Зарегистрируйтесь через бота."})
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
