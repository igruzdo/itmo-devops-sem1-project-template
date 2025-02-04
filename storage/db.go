package storage

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// ConnectDB устанавливает соединение с PostgreSQL
func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", "user=validator password=val1dat0r dbname=project-sem-1 sslmode=disable")
	if err != nil {
		log.Println("Ошибка подключения к БД:", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Println("БД недоступна:", err)
		return nil, err
	}

	log.Println("Подключение к БД успешно!")
	return db, nil
}