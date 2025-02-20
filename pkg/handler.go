package pkg

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) HandlePrices(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request on /api/v0/prices", r.Method)

	switch r.Method {
	case http.MethodPost:
		h.handleUpload(w, r)
	case http.MethodGet:
		h.handleDownload(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleUpload(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting file upload...")

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read file bytes: %v", err)
		http.Error(w, "Failed to read file bytes", http.StatusInternalServerError)
		return
	}

	log.Printf("File size: %d bytes", len(fileBytes))

	zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		log.Printf("Invalid ZIP file: %v", err)
		http.Error(w, "Invalid ZIP file", http.StatusBadRequest)
		return
	}

	var records [][]string
	for _, f := range zipReader.File {
		if f.Name == "data.csv" {
			log.Println("Processing CSV file...")

			csvFile, err := f.Open()
			if err != nil {
				log.Printf("Failed to open CSV: %v", err)
				http.Error(w, "Failed to open CSV", http.StatusInternalServerError)
				return
			}
			defer csvFile.Close()

			reader := csv.NewReader(csvFile)
			_, err = reader.Read()
			if err != nil {
				log.Printf("Failed to read CSV header: %v", err)
				http.Error(w, "Failed to read CSV header", http.StatusInternalServerError)
				return
			}

			records, err = reader.ReadAll()
			if err != nil {
				log.Printf("Failed to read CSV rows: %v", err)
				http.Error(w, "Failed to read CSV rows", http.StatusInternalServerError)
				return
			}
		}
	}

	tx, err := h.db.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	for _, record := range records {
		price, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			log.Printf("Invalid price format: %s", record[3])
			http.Error(w, fmt.Sprintf("Invalid price format: %s", record[3]), http.StatusBadRequest)
			tx.Rollback()
			return
		}

		_, err = tx.Exec(`INSERT INTO prices (create_date, name, category, price) 
							VALUES ($1::date, $2, $3, $4::numeric)`,
			record[4], record[1], record[2], price)

		if err != nil {
			log.Printf("Failed to insert data: %v", err)
			http.Error(w, "Failed to insert data", http.StatusInternalServerError)
			tx.Rollback()
			return
		}
	}

	var totalItems, totalCategories int
	var totalPrice float64

	err = tx.QueryRow("SELECT COUNT(*), COUNT(DISTINCT category), SUM(price) FROM prices").Scan(&totalItems, &totalCategories, &totalPrice)
	if err != nil {
		log.Printf("Failed to calculate totals: %v", err)
		http.Error(w, "Failed to calculate totals", http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_items":      totalItems,
		"total_categories": totalCategories,
		"total_price":      totalPrice,
	})

	log.Printf("Upload completed successfully: %d items, %d categories, total price: %.2f",
		totalItems, totalCategories, totalPrice)
}

func (h *Handler) handleDownload(w http.ResponseWriter, _ *http.Request) {
	log.Println("Starting data download...")

	rows, err := h.db.Query("SELECT TO_CHAR(create_date, 'YYYY-MM-DD'), name, category, price FROM prices")
	if err != nil {
		log.Printf("Failed to query database: %v", err)
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records [][]string
	for rows.Next() {
		var createDate, name, category string
		var price float64

		if err := rows.Scan(&createDate, &name, &category, &price); err != nil {
			log.Printf("Failed to process row: %v", err)
			http.Error(w, "Failed to process row", http.StatusInternalServerError)
			return
		}

		records = append(records, []string{createDate, name, category, fmt.Sprintf("%.2f", price)})
	}

	var csvBuffer bytes.Buffer
	writer := csv.NewWriter(&csvBuffer)
	writer.Write([]string{"create_date", "name", "category", "price"})
	writer.WriteAll(records)
	writer.Flush()

	if err := writer.Error(); err != nil {
		log.Printf("Failed to write CSV: %v", err)
		http.Error(w, "Failed to write CSV", http.StatusInternalServerError)
		return
	}

	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)

	file, err := zipWriter.Create("data.csv")
	if err != nil {
		log.Printf("Failed to create CSV inside ZIP: %v", err)
		http.Error(w, "Failed to create CSV inside ZIP", http.StatusInternalServerError)
		return
	}

	_, err = file.Write(csvBuffer.Bytes())
	if err != nil {
		log.Printf("Failed to write CSV to ZIP: %v", err)
		http.Error(w, "Failed to write CSV to ZIP", http.StatusInternalServerError)
		return
	}

	zipWriter.Close()

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=response.zip")
	w.Write(zipBuffer.Bytes())

	log.Println("Download completed successfully")
}
