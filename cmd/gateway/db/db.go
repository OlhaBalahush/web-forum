package db

import (
	"database/sql"

	"github.com/OlhaBalahush/web-forum/cmd/gateway/helpers"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func OpenDatabase() (*Database, error) {
	var db *sql.DB
	var err error
	if !*helpers.Dockerize {
		db, err = sql.Open("sqlite3", "db/src/database.db")
	} else {
		db, err = sql.Open("sqlite3", "/data/db/database.db")
	}
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}
