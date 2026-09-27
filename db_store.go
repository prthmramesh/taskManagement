package main

import (
	"database/sql"
	"fmt"
	"log"
)

var db *sql.DB

func initDB() {
	connStr := "host=localhost port=5432 user=postgres password=password123 dbname=task_management sslmode=disable"

	var err error

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database!")

}

func insertUser(username, password, email string) error {
	sqlStatement := `INSERT INTO users (username, password_hash, email, created_at) VALUES ($1, $2, $3, NOW())`
	_, err := db.Exec(sqlStatement, username, password, email)
	return err
}
