package db

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type database struct {
	db *sql.DB
}

func Init(dbHost, dbPort, dbUser, dbPassword, dbName string) (*database, error) {
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	instance := database{
		db: db,
	}

	if err := instance.checkTable(); err != nil {
		return nil, err
	}

	return &instance, nil
}

func (r *database) checkTable() error {
	rows, err := r.db.Query("SELECT * FROM images LIMIT 1")
	if err != nil {
		c, ok := err.(*pq.Error)
		if ok && c.Code != "42P01" {
			return err
		}
		_, err = r.db.Exec(`
			CREATE TABLE images (
				id SERIAL PRIMARY KEY,
				data TEXT NOT NULL
			)
		`)
		if err != nil {
			return err
		}
	} else {
		rows.Close()
	}
	return nil
}

func (r *database) GetDb() *sql.DB {
	return r.db
}

func (r *database) Close() {
	defer r.db.Close()
}
