package handlers

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"project_sem/db"
	"strconv"
	"strings"
)

func HandlePostPrices(w http.ResponseWriter, r *http.Request) {
	dbConnection, dbConerr := db.GetPostgreConnection()

	if dbConerr != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", dbConerr)
	}
	
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
	tx, err := dbConnection.Begin()
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