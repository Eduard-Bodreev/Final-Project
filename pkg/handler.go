package pkg

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

func HandlePrices(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request on /api/v0/prices", r.Method)

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
	log.Println("Starting file upload...")

	db, err := InitDB()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer db.Close()

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

	var totalItems int
	categorySet := make(map[string]struct{})
	var totalPrice float64

	for _, f := range zipReader.File {
		log.Printf("Found file in ZIP: %s", f.Name)
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

			header, err := reader.Read()
			if err != nil {
				log.Printf("Failed to read CSV header: %v", err)
				http.Error(w, "Failed to read CSV header", http.StatusInternalServerError)
				return
			}
			log.Printf("CSV header: %v", header)

			tx, err := db.Begin()
			if err != nil {
				log.Printf("Failed to start transaction: %v", err)
				http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
				return
			}

			for {
				record, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("Failed to read CSV row: %v", err)
					http.Error(w, "Failed to read CSV", http.StatusInternalServerError)
					tx.Rollback()
					return
				}

				log.Printf("Processing row: %v", record)

				id, err := strconv.Atoi(record[0])
				if err != nil {
					log.Printf("Invalid ID format: %s", record[0])
					http.Error(w, fmt.Sprintf("Invalid ID format: %s", record[0]), http.StatusBadRequest)
					tx.Rollback()
					return
				}

				price, err := strconv.ParseFloat(record[3], 64)
				if err != nil {
					log.Printf("Invalid price format: %s", record[3])
					http.Error(w, fmt.Sprintf("Invalid price format: %s", record[3]), http.StatusBadRequest)
					tx.Rollback()
					return
				}

				_, err = tx.Exec(`INSERT INTO prices (id, create_date, name, category, price) 
								  VALUES ($1, $2::date, $3, $4, $5::numeric)`,
					id, record[4], record[1], record[2], fmt.Sprintf("%.2f", price))

				if err != nil {
					log.Printf("Failed to insert data: %v", err)
					http.Error(w, "Failed to insert data", http.StatusInternalServerError)
					tx.Rollback()
					return
				}

				totalItems++
				totalPrice += price
				categorySet[record[2]] = struct{}{}
			}

			if err := tx.Commit(); err != nil {
				log.Printf("Failed to commit transaction: %v", err)
				http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
				return
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

	log.Printf("Upload completed successfully: %d items, %d categories, total price: %.2f",
		totalItems, totalCategories, totalPrice)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting data download...")

	db, err := InitDB()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, TO_CHAR(create_date, 'YYYY-MM-DD'), name, category, price FROM prices")
	if err != nil {
		log.Printf("Failed to query database: %v", err)
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var csvBuffer bytes.Buffer
	writer := csv.NewWriter(&csvBuffer)

	writer.Write([]string{"id", "create_date", "name", "category", "price"})

	for rows.Next() {
		var id int
		var createDate, name, category string
		var price float64

		err := rows.Scan(&id, &createDate, &name, &category, &price)
		if err != nil {
			log.Printf("Failed to process row: %v", err)
			http.Error(w, "Failed to process row", http.StatusInternalServerError)
			return
		}

		writer.Write([]string{strconv.Itoa(id), createDate, name, category, fmt.Sprintf("%.2f", price)})
	}
	writer.Flush()

	if err := rows.Err(); err != nil {
		log.Printf("Error reading rows: %v", err)
		http.Error(w, "Error reading rows", http.StatusInternalServerError)
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
