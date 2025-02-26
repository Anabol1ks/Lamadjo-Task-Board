package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/Anabol1ks/Lamadjo-Task-Board/docs"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/auth"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/models"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/storage"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/team"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @Title Сервис для контроля задачами и встречами команды
func main() {
	key := os.Getenv("DB_HOST")
	if key == "" {
		fmt.Println("Используется данные из .env")
		if err := godotenv.Load(); err != nil {
			log.Fatal("Ошибка получения .env")
		}
	}

	storage.ConnectDatabase()

	if err := storage.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("Ошибка миграции пользователей: ", err.Error())
	}
	if err := storage.DB.AutoMigrate(&models.Team{}, &models.Task{}, &models.Meeting{}, &models.Notification{}); err != nil {
		log.Fatal("Ошибка миграции остальных моделей: ", err.Error())
	}

	// Инициализация бота
	//

	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Эндпоинты для авторизации
	r.POST("/auth", auth.RegisterHandler)
	r.GET("/auth", auth.CheckAuthHandler)
	//

	// Эндпоинты для управления командами
	r.POST("/team", team.CreateTeamHandler)
	r.POST("/team/join", team.JoinTeamHandler)
	r.GET("/team/my", team.GetMyTeamHandler)
	r.GET("/team/invite", team.GetLinkTeamHandler)
	r.GET("/team/members", team.GetMembersTeam)
	r.GET("/team/leave", team.LeaveMemberTeamHandler)
	r.PUT("/team", team.ChangeTeamHandler)
	//

	// Эндпоинты для управления участниками команды
	//

	// Эндпоинты задач
	//

	// Эндпоинты встреч
	//

	// Модуль уведомлений
	//

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
