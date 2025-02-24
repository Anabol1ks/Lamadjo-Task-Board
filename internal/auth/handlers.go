package auth

import (
	"log"
	"net/http"

	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/models"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/storage"
	"github.com/gin-gonic/gin"
)

type RegisterInput struct {
	TelegramID string `json:"telegram_id" binding:"required"` // Уникальный идентификатор Telegram
	Name       string `json:"name"`
	Role       string `json:"role" binding:"required,oneof=manager member"` // "manager" или "member"
}

// AuthHandler godoc
// @Summary Регистрация пользователя
// @Description Регистрация пользователя с помощью уникального telegram_id
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RegisterInput true "Данные пользователя для регистрации"
// @Success 200 {object} response.SuccessResponse "Успешная регистрация"
// @Failure 400 {object} response.ErrorResponse "Ошибка валидации"
// @Failure 409 {object} response.ErrorResponse "Пользователь уже зарегистрирован"
// @Failure 507 {object} response.ErrorResponse "Не удалось создать пользователя"
// @Router /auth [post]
func RegisterHandler(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.User
	if err := storage.DB.Where("telegram_id = ?", input.TelegramID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Пользователь уже зарегистрирован"})
		return
	}

	user := models.User{
		TelegramID: input.TelegramID,
		Name:       input.Name,
		Role:       input.Role,
	}

	if err := storage.DB.Create(&user).Error; err != nil {
		log.Println("Не удалось создать пользователя", err.Error())
		c.JSON(http.StatusInsufficientStorage, gin.H{"error": "Не удалось создать пользователя"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Успешная регистрация"})
}

// CheckAuthHandler godoc
// @Summary Проверка авторизации пользователя
// @Description Проверяет, зарегистрирован ли пользователь по telegram_id. Если пользователь найден, возвращает его данные, иначе – сообщение об ошибке.
// @Tags auth
// @Accept json
// @Produce json
// @Param telegram_id query string true "Уникальный идентификатор Telegram"
// @Success 200 {object} RegisterInput "Данные пользователя"
// @Failure 400 {object} response.ErrorResponse "Ошибка telegram_id is required"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Router /auth [get]
func CheckAuthHandler(c *gin.Context) {
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

	c.JSON(http.StatusOK, user)
}
