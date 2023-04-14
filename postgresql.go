package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func connectToDB() *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.name"),
		viper.GetString("database.sslmode"))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error("Cannot connect to database: \n", err)
		return nil
	}
	if err = db.Ping(); err != nil {
		db.Close()
		log.Error("Database connection interrupted: \n", err)
		return nil
	}
	return db
}

func readFromDB(db *sql.DB, query string) (*sql.Row, error) {
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	return row, nil
}

func writeToDB(db *sql.DB, query string, values ...interface{}) error {
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(values...)
	if err != nil {
		return err
	}

	return nil
}
