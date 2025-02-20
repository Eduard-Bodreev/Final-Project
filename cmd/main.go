package main

import (
	"log"
	"net/http"

	"github.com/Eduard-Bodreev/Final-Project/pkg"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	db, err := pkg.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	handler := pkg.NewHandler(db)

	http.HandleFunc("/api/v0/prices", handler.HandlePrices)

	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
