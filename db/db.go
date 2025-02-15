package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

type DbConnectionSettings struct {
	host string
	port string
	user string
	password string
	name string
}

var connection *sql.DB

func getSettings() (*DbConnectionSettings) {

	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "validator")
	password := getEnv("POSTGRES_PASSWORD", "val1dat0r")
	name := getEnv("POSTGRES_DB", "project-sem-1")

	settings := DbConnectionSettings{
		host: host,
		port: port,
		user: user,
		password: password,
		name: name,
	}

	return &settings
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}


func GetPostgreConnection() (*sql.DB, error) {

	if connection != nil {
		return connection, nil
	}

	settings := getSettings();

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		settings.host, settings.port, settings.user, settings.password, settings.name)

	log.Printf("Подключаемся к базе данных: %s", connStr)

	var err error
	connection, err = sql.Open("postgres", connStr)

	return connection, err
}