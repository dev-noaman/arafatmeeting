//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to PostgreSQL default database
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Create mini_meeting database
	_, err = db.Exec("CREATE DATABASE mini_meeting")
	if err != nil {
		log.Printf("Database may already exist: %v", err)
	} else {
		fmt.Println("Database 'mini_meeting' created successfully")
	}
}
