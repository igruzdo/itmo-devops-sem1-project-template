package main

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "validator"
	password = "val1dat0r"
	dbname   = "project-sem-1"
)

var db *sql.DB

func main() {
	// Подключение к базе данных
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Инициализация Gin
	r := gin.Default()

	// Эндпоинт POST /api/v0/prices
	r.POST("/api/v0/prices", handlePost)

	// Эндпоинт GET /api/v0/prices
	r.GET("/api/v0/prices", handleGet)

	// Запуск сервера
	r.Run(":8080")
}

func handlePost(c *gin.Context) {
	// Чтение архива из тела запроса
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Открытие архива
	zipFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer zipFile.Close()

	// Чтение файла data.csv из архива
	reader, err := zip.NewReader(zipFile, file.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var totalItems int
	var totalPrice float64
	categories := make(map[string]bool)

	for _, f := range reader.File {
		if f.Name == "data.csv" {
			rc, err := f.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer rc.Close()

			csvReader := csv.NewReader(rc)
			for {
				record, err := csvReader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				// Парсинг данных
				productID := record[0]
				createdAt, _ := time.Parse("2006-01-02", record[1])
				productName := record[2]
				category := record[3]
				price, _ := strconv.ParseFloat(record[4], 64)

				// Вставка данных в базу
				_, err = db.Exec("INSERT INTO prices (product_id, created_at, product_name, category, price) VALUES ($1, $2, $3, $4, $5)",
					productID, createdAt, productName, category, price)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				// Обновление статистики
				totalItems++
				categories[category] = true
				totalPrice += price
			}
		}
	}

	// Возврат JSON с результатами
	c.JSON(http.StatusOK, gin.H{
		"total_items":     totalItems,
		"total_categories": len(categories),
		"total_price":     totalPrice,
	})
}

func handleGet(c *gin.Context) {
	// Выборка данных из базы
	rows, err := db.Query("SELECT product_id, created_at, product_name, category, price FROM prices")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Создание временного файла CSV
	csvFile, err := os.CreateTemp("", "data.csv")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(csvFile.Name())

	csvWriter := csv.NewWriter(csvFile)
	for rows.Next() {
		var productID, productName, category string
		var createdAt time.Time
		var price float64
		err = rows.Scan(&productID, &createdAt, &productName, &category, &price)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		csvWriter.Write([]string{productID, createdAt.Format("2006-01-02"), productName, category, fmt.Sprintf("%.2f", price)})
	}
	csvWriter.Flush()

	// Создание zip-архива
	zipFile, err := os.CreateTemp("", "data.zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(zipFile.Name())

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Добавление CSV файла в архив
	fileInZip, err := zipWriter.Create("data.csv")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = io.Copy(fileInZip, csvFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возврат архива
	c.File(zipFile.Name())
}