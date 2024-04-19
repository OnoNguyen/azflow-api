package postgresql

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var Db *sql.DB

func InitDB() {
	connStr := "user=azflow-admin dbname=azflowcore password=abcd1234 host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	Db = db
}

func CloseDB() error {
	return Db.Close()
}

func Migrate() {
	if err := Db.Ping(); err != nil {
		panic(err)
	}

	driverURL := "postgres://postgres:abcd1234@localhost:5432/azflowcore?sslmode=disable"
	m, err := migrate.New("file://internal/pkg/db/migrations/postgresql", driverURL)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(err)
	}
}
