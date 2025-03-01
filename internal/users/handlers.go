package users

import (
	"net/http"

	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/models"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/response"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/storage"
	"github.com/gin-gonic/gin"
)

// @Summary Получение информации о пользователе
// @Description Получает информацию о пользователе по его Telegram ID.
// @Tags users
// @Accept json
// @Produce json
// @Param telegram_id query string true "Telegram ID of the user"
// @Success 200 {object} response.UserInfoResponse "Информация о пользователе"
// @Failure 400 {object} response.ErrorResponse "Ошибка валидации или отсутствует telegram_id"
// @Failure 401 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} response.ErrorResponse "Ошибка создания команды"
// @Router /user [get]
func GetMyUser(c *gin.Context) {
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

	info := response.UserInfoResponse{
		Name:     user.Name,
		Role:     user.Role,
		TeamName: "Нет команды",
	}

	if user.TeamID == nil {
		var team models.Team
		if err := storage.DB.Where("id = ?", user.TeamID).First(&team).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить информацию о команде"})
			return
		}

		info.TeamName = team.Name
	}

	c.JSON(http.StatusOK, info)
}
