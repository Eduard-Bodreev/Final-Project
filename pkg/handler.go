package pkg

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func HandlePrices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleUpload(w, r)
	case http.MethodGet:
		handleDownload(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	db, err := InitDB()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file bytes", http.StatusInternalServerError)
		return
	}

	zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		http.Error(w, "Invalid ZIP file", http.StatusBadRequest)
		return
	}

	var totalItems int
	categorySet := make(map[string]struct{})
	var totalPrice float64

	for _, f := range zipReader.File {
		if f.Name == "data.csv" {
			csvFile, err := f.Open()
			if err != nil {
				http.Error(w, "Failed to open CSV", http.StatusInternalServerError)
				return
			}
			defer csvFile.Close()

			reader := csv.NewReader(csvFile)

			_, err = reader.Read()
			if err != nil {
				http.Error(w, "Failed to read CSV header", http.StatusInternalServerError)
				return
			}

			for {
				record, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					http.Error(w, "Failed to read CSV", http.StatusInternalServerError)
					return
				}

				id, err := strconv.Atoi(record[0])
				if err != nil {
					http.Error(w, fmt.Sprintf("Invalid ID format: %s", record[0]), http.StatusBadRequest)
					return
				}

				price, err := strconv.ParseFloat(record[3], 64)
				if err != nil {
					http.Error(w, fmt.Sprintf("Invalid price format: %s", record[3]), http.StatusBadRequest)
					return
				}

				_, err = db.Exec(`INSERT INTO prices (id, created_at, name, category, price) 
								  VALUES ($1, $2::date, $3, $4, $5::numeric)`,
					id, record[4], record[1], record[2], fmt.Sprintf("%.2f", price))

				if err != nil {
					errorMessage := fmt.Sprintf("Failed to insert data: %v", err)
					http.Error(w, errorMessage, http.StatusInternalServerError)
					return
				}

				totalItems++
				totalPrice += price
				categorySet[record[2]] = struct{}{}
			}
		}
	}

	totalCategories := len(categorySet)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_items":      totalItems,
		"total_categories": totalCategories,
		"total_price":      totalPrice,
	})
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	db, err := InitDB()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, TO_CHAR(created_at, 'YYYY-MM-DD'), name, category, price FROM prices")
	if err != nil {
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	file, err := zipWriter.Create("data.csv")
	if err != nil {
		http.Error(w, "Failed to create CSV", http.StatusInternalServerError)
		return
	}

	writer := csv.NewWriter(file)
	for rows.Next() {
		var id int
		var createdAt, name, category string
		var price float64

		err := rows.Scan(&id, &createdAt, &name, &category, &price)
		if err != nil {
			http.Error(w, "Failed to process row", http.StatusInternalServerError)
			return
		}

		writer.Write([]string{strconv.Itoa(id), createdAt, name, category, fmt.Sprintf("%.2f", price)})
	}
	writer.Flush()
}
