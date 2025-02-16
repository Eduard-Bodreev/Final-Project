package pkg

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "validator"
	password = "val1dat0r"
	dbname   = "project-sem-1"
)

func InitDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("Database is ready!")
			return db, nil
		}
		log.Printf("⏳ Database not ready, retrying in 2s... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("database not ready after multiple attempts: %v", err)
}
