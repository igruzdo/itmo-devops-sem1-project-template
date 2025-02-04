package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"project_sem/handlers"
)

func main() {
	// Подключение к базе данных
	db, err := sql.Open("postgres", "user=validator password=val1dat0r dbname=project-sem-1 sslmode=disable")
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}
	defer db.Close()

	// Проверка соединения
	if err := db.Ping(); err != nil {
		log.Fatal("База данных недоступна:", err)
	}

	// Создаём роутер
	router := gin.Default()

	// Регистрируем эндпоинты
	router.POST("/api/v0/prices", handlers.UploadPrices(db))
	router.GET("/api/v0/prices", handlers.DownloadPrices(db))

	// Запускаем сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Сервер запущен на порту", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}