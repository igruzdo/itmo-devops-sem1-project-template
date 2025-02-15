package main

import (
	"log"
	"net/http"
	"project_sem/db"
	"project_sem/handlers"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)



func main() {

	dbConnection, err := db.GetPostgreConnection()

	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer dbConnection.Close()

	// Проверка соединения
	err = dbConnection.Ping()
	if err != nil {
		log.Fatalf("База данных недоступна: %v", err)
	}

	log.Println("Соединение с базой данных успешно установлено!")

	// Настройка маршрутов
	router := mux.NewRouter()
	router.HandleFunc("/api/v0/prices", handlers.HandlePostPrices).Methods("POST")
	router.HandleFunc("/api/v0/prices", handlers.HandleGetPrices).Methods("GET")

	// Запуск сервера
	log.Println("Сервер запущен на порту 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}