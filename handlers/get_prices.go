package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"project_sem/db"
	"strconv"
)

func HandleGetPrices(w http.ResponseWriter, r *http.Request) {
	dbConnection, dbConerr := db.GetPostgreConnection()

	if dbConerr != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", dbConerr)
	}
	
	rows, err := dbConnection.Query("SELECT id, name, category, price, create_date FROM prices")
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