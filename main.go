package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/Anabol1ks/Lamadjo-Task-Board/docs"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/auth"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/meetings"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/models"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/storage"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/tasks"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/team"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/users"
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
	if err := storage.DB.AutoMigrate(&models.Team{}, &models.Task{}, &models.Meeting{}, &models.Notification{}, &models.Room{}, &models.InviteLink{}); err != nil {
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

	r.GET("/user", users.GetMyUser)

	teamGroup := r.Group("/team")
	// Эндпоинты для управления командами
	{
		teamGroup.POST("", team.CreateTeamHandler)
		teamGroup.POST("/join", team.JoinTeamHandler)
		teamGroup.GET("/my", team.GetMyTeamHandler)
		teamGroup.GET("/invite", team.GetLinkTeamHandler)
		teamGroup.GET("/leave", team.LeaveMemberTeamHandler)
		teamGroup.PUT("", team.ChangeTeamHandler)
		teamGroup.DELETE("", team.DeleteTeamHandler)
		//

		// Эндпоинты для управления участниками команды
		teamGroup.GET("/members", team.GetMembersTeam)
		teamGroup.GET("/kick", team.KickMemberTeamHandler)
		//
	}

	// Эндпоинты задач
	tasksGroup := r.Group("/tasks")
	{
		tasksGroup.POST("", tasks.CreateTaskHandlres)
		tasksGroup.GET("", tasks.GetTasksHandlres)
		tasksGroup.DELETE("/:id", tasks.DeleteTaskHandler)
		tasksGroup.PUT("/:id/status", tasks.UpdateTaskStatusHandler)
	}
	//

	meetingsGroup := r.Group("/meetings")
	{
		meetingsGroup.POST("/", meetings.CreateMeetingHandler)
		meetingsGroup.GET("/available-slots", meetings.GetAvailableTimeSlotsHandler)
		meetingsGroup.DELETE("/:id", meetings.DeleteMeetingHandler)
		meetingsGroup.GET("/my", meetings.GetMyMeeting)
	}
	// Эндпоинты встреч
	//

	// Модуль уведомлений
	//

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
