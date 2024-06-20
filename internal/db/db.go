package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

var DB *sql.DB

const connStr = "user=postgres dbname=Positiv sslmode=disable port=5433"

func init() {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
}
