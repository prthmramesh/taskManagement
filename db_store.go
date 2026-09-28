package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

var db *sql.DB

func initDB() {

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		log.Fatal("DB_PASSWORD is not set")
	}
	connStr := fmt.Sprintf(
		"host=localhost port=5432 user=postgres password='%s' dbname=task_management sslmode=disable",
		dbPassword,
	)

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

func getUserByUsername(username string) (*User, error) {
	sqlStatement := `SELECT username, password_hash, email, id FROM users WHERE username=$1`
	row := db.QueryRow(sqlStatement, username)
	user := &User{}
	err := row.Scan(&user.Username, &user.Password, &user.Email, &user.ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
