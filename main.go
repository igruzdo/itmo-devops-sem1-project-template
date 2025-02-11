package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	// Подключение к PostgreSQL
	dbHost := getEnv("POSTGRES_HOST", "localhost")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "validator")
	dbPassword := getEnv("POSTGRES_PASSWORD", "val1dat0r")
	dbName := getEnv("POSTGRES_DB", "project-sem-1")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Подключаемся к базе данных: %s", connStr)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	// Проверка соединения
	err = db.Ping()
	if err != nil {
		log.Fatalf("База данных недоступна: %v", err)
	}

	log.Println("Соединение с базой данных успешно установлено!")

	// Настройка маршрутов
	router := mux.NewRouter()
	router.HandleFunc("/api/v0/prices", handlePostPrices).Methods("POST")
	router.HandleFunc("/api/v0/prices", handleGetPrices).Methods("GET")

	// Запуск сервера
	log.Println("Сервер запущен на порту 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// Функция для получения значения переменной окружения с возможностью задания значения по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func handlePostPrices(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что запрос содержит файл
	err := r.ParseMultipartForm(10 << 20) // Максимальный размер: 10 MB
	if err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	// Получаем файл из формы
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка загрузки файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Загружен файл: %s\n", handler.Filename)

	// Читаем содержимое файла в память
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Ошибка чтения файла", http.StatusInternalServerError)
		return
	}

	// Открываем ZIP-архив из памяти
	zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		http.Error(w, "Ошибка разархивации файла", http.StatusBadRequest)
		log.Printf("Ошибка разархивации файла: %v", err)
		return
	}

	var csvFile string
	for _, f := range zipReader.File {
		if strings.HasSuffix(f.Name, ".csv") {
			csvFile = f.Name
			break
		}
	}

	if csvFile == "" {
		http.Error(w, "CSV-файл не найден в архиве", http.StatusBadRequest)
		return
	}

	// Начинаем транзакцию
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Ошибка начала транзакции", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Читаем содержимое CSV-файла
	var totalItems int
	categories := make(map[string]struct{})

	for _, f := range zipReader.File {
		if f.Name == csvFile {
			rc, err := f.Open()
			if err != nil {
				http.Error(w, "Ошибка открытия CSV-файла", http.StatusInternalServerError)
				return
			}
			defer rc.Close()

			reader := csv.NewReader(bufio.NewReader(rc))
			reader.Read() // Пропускаем заголовок
			for {
				record, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					http.Error(w, "Ошибка чтения CSV-файла", http.StatusInternalServerError)
					return
				}

				// Проверка на корректность данных
				if len(record) < 5 {
					http.Error(w, "Некорректный формат записи в CSV-файле", http.StatusBadRequest)
					return
				}

				name := record[1]
				category := record[2]
				price, err := strconv.ParseFloat(record[3], 64)
				if err != nil {
					http.Error(w, "Некорректное значение цены", http.StatusBadRequest)
					return
				}
				createDate := record[4]

				// Сохраняем в базу данных
				_, err = tx.Exec(
					"INSERT INTO prices (name, category, price, create_date) VALUES ($1, $2, $3, $4)",
					name, category, price, createDate,
				)
				if err != nil {
					http.Error(w, "Ошибка записи в базу данных", http.StatusInternalServerError)
					return
				}

				totalItems++
				categories[category] = struct{}{}
			}
		}
	}

	// Подсчет total_price и total_categories по всей таблице в рамках транзакции
	var totalPrice float64
	var totalCategories int

	// Запрос для подсчета суммарной стоимости и количества категорий
	err = tx.QueryRow("SELECT SUM(price), COUNT(DISTINCT category) FROM prices").Scan(&totalPrice, &totalCategories)
	if err != nil {
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		return
	}

	// Завершаем транзакцию
	if err := tx.Commit(); err != nil {
		http.Error(w, "Ошибка завершения транзакции", http.StatusInternalServerError)
		return
	}

	// Формируем JSON-ответ
	response := map[string]interface{}{
		"total_items":      totalItems,
		"total_categories": totalCategories,
		"total_price":      totalPrice,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetPrices(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, category, price, create_date FROM prices")
	if err != nil {
		http.Error(w, "Ошибка получения данных из базы", http.StatusInternalServerError)
		log.Printf("Ошибка базы данных: %v", err)
		return
	}
	defer rows.Close()

	// Создаем CSV-файл в памяти
	var csvBuffer bytes.Buffer
	writer := csv.NewWriter(&csvBuffer)
	writer.Write([]string{"id", "name", "category", "price", "create_date"})

	for rows.Next() {
		var id int
		var name, category, createDate string
		var price float64

		err := rows.Scan(&id, &name, &category, &price, &createDate)
		if err != nil {
			http.Error(w, "Ошибка чтения данных из базы", http.StatusInternalServerError)
			log.Printf("Ошибка чтения строки: %v", err)
			return
		}

		writer.Write([]string{
			strconv.Itoa(id), name, category, fmt.Sprintf("%.2f", price), createDate,
		})
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Ошибка при обработке строк", http.StatusInternalServerError)
		log.Printf("Ошибка при обработке строк: %v", err)
		return
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		http.Error(w, "Ошибка записи в CSV", http.StatusInternalServerError)
		log.Printf("Ошибка записи в CSV: %v", err)
		return
	}

	// Создаем ZIP-файл в памяти
	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)
	fileWriter, err := zipWriter.Create("data.csv")
	if err != nil {
		http.Error(w, "Ошибка добавления файла в ZIP", http.StatusInternalServerError)
		log.Printf("Ошибка записи в ZIP: %v", err)
		return
	}

	_, err = fileWriter.Write(csvBuffer.Bytes())
	if err != nil {
		http.Error(w, "Ошибка копирования данных в ZIP", http.StatusInternalServerError)
		log.Printf("Ошибка копирования данных в ZIP: %v", err)
		return
	}

	if err := zipWriter.Close(); err != nil {
		http.Error(w, "Ошибка закрытия ZIP-файла", http.StatusInternalServerError)
		log.Printf("Ошибка закрытия ZIP-файла: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"data.zip\"")
	w.Write(zipBuffer.Bytes())
}