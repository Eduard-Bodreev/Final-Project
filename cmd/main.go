package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Eduard-Bodreev/Final-Project/pkg"
)

func main() {
	db, err := pkg.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/api/v0/prices", pkg.HandlePrices)

	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
