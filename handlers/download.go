package handlers

import (
	"database/sql"
	"encoding/csv"
	"net/http"
	"os"
	"strconv"

	"project_sem/utils" // Импорт утилит

	"github.com/gin-gonic/gin"
)

// DownloadPrices выгружает данные из БД и архивирует их в ZIP
func DownloadPrices(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT product_id, created_at, name, category, price FROM prices")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
			return
		}
		defer rows.Close()

		// Создаём временный CSV-файл
		tempFile, err := os.CreateTemp("", "data-*.csv")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp file"})
			return
		}
		defer os.Remove(tempFile.Name()) // Удаляем после отправки

		writer := csv.NewWriter(tempFile)
		for rows.Next() {
			var id int
			var createdAt, name, category string
			var price float64
			if err := rows.Scan(&id, &createdAt, &name, &category, &price); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read data"})
				return
			}

			writer.Write([]string{
				strconv.Itoa(id),
				createdAt,
				name,
				category,
				strconv.FormatFloat(price, 'f', 2, 64),
			})
		}
		writer.Flush()
		tempFile.Close()

		// Архивируем CSV в ZIP
		zipPath := tempFile.Name() + ".zip"
		if err := utils.Zip(tempFile.Name(), zipPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ZIP"})
			return
		}

		// Отправляем ZIP-файл клиенту
		c.File(zipPath)
	}
}