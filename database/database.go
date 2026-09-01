package database

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() error {
	db, err := sql.Open("sqlite", "database.db")
	if err != nil {
		return errors.New("Error opening database")
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("pinging database %w", err)
	}
	return nil
}

func CreateTables() error {
	query := `
	CREATE TABLE IF NOT EXIST users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	date_of_birth TEXT NOT NULL,
	role TEXT NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS applications (
	id INTERGER PRIMARY KEY AUTOINCREMENT, 
	user_id INTEGER NOT NULL

	);

	CREATE TABLE IF NOT EXISTS companies (
	id INTERGER PRIMARY KEY AUTOINCREMENT,
	user_id INTERGER NOT NULL,
	company_name TEXT NOT NULL, 
	description TEXT NOT NULL, 
	website TEXT NOT NULL, 
	location TEXT NOT NULL,
	created_at DATETIME NOT NULL
	);
	`
}
