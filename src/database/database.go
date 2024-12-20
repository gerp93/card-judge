package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var database *sql.DB

func CreateDatabaseConnection() (*sql.DB, error) {
	// get connection string
	userName := os.Getenv("CARD_JUDGE_SQL_USER")
	userPassword := os.Getenv("CARD_JUDGE_SQL_PASSWORD")
	serverHost := os.Getenv("CARD_JUDGE_SQL_HOST")
	databaseName := os.Getenv("CARD_JUDGE_SQL_DATABASE")
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", userName, userPassword, serverHost, databaseName)

	// open database connection
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Println(err)
		return db, errors.New("failed to open database connection")
	}

	// set global variable for database connection
	database = db

	// ping to test connection
	err = db.Ping()
	if err != nil {
		log.Println(err)
		db.Close()
		return database, errors.New("failed to ping database")
	}

	return database, nil
}

func RunFile(filePath string) error {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Println(err)
		return errors.New("failed to read file")
	}

	return execute(string(bytes))
}

func query(sqlString string, params ...any) (*sql.Rows, error) {
	statement, err := database.Prepare(sqlString)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	rows, err := statement.Query(params...)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to query statement in database")
	}

	return rows, nil
}

func execute(sqlString string, params ...any) error {
	statement, err := database.Prepare(sqlString)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	_, err = statement.Exec(params...)
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}
