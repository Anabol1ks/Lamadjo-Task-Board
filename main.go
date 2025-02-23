package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/models"
	"github.com/Anabol1ks/Lamadjo-Task-Board/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	key := os.Getenv("DB_HOST")
	if key == "" {
		fmt.Println("Используется данные из .env")
		if err := godotenv.Load(); err != nil {
			log.Fatal("Ошибка получения .env")
		}
	}

	storage.ConnectDatabase()

	if err := storage.DB.AutoMigrate(
		&models.User{},         // сначала создаём таблицу пользователей
		&models.Team{},         // затем таблица команд, которая ссылается на пользователей
		&models.Task{},         // далее таблица задач
		&models.Meeting{},      // затем таблица встреч
		&models.Notification{}, // и, наконец, таблица уведомлений
	); err != nil {
		log.Fatal("Ошибка миграции: ", err.Error())
	}

	r := gin.Default()

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
