package handlers

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// UploadPrices обрабатывает загрузку ZIP-архива с CSV
func UploadPrices(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
			return
		}
		defer file.Close()

		tempFile, err := os.CreateTemp("", "upload-*.zip")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp file"})
			return
		}
		defer os.Remove(tempFile.Name())

		_, err = io.Copy(tempFile, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save temp file"})
			return
		}
		tempFile.Close()

		r, err := zip.OpenReader(tempFile.Name())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unzip"})
			return
		}
		defer r.Close()

		var csvFile *zip.File
		for _, f := range r.File {
			if strings.HasSuffix(f.Name, ".csv") {
				csvFile = f
				break
			}
		}
		if csvFile == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file not found in ZIP"})
			return
		}

		rc, err := csvFile.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open CSV"})
			return
		}
		defer rc.Close()

		reader := csv.NewReader(rc)
		reader.FieldsPerRecord = 5

		tx, _ := db.Begin()
		stmt, _ := tx.Prepare("INSERT INTO prices (product_id, created_at, name, category, price) VALUES ($1, $2, $3, $4, $5)")
		defer stmt.Close()

		totalItems, totalPrice := 0, 0.0
		categories := make(map[string]struct{})

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}

			id, _ := strconv.Atoi(record[0])
			price, _ := strconv.ParseFloat(record[4], 64)

			stmt.Exec(id, record[1], record[2], record[3], price)

			totalItems++
			categories[record[3]] = struct{}{}
			totalPrice += price
		}

		tx.Commit()

		c.JSON(http.StatusOK, gin.H{
			"total_items":     totalItems,
			"total_categories": len(categories),
			"total_price":     totalPrice,
		})
	}
}