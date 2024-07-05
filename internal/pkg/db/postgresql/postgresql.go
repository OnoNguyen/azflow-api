package database

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"os"
)

var Db *pgx.Conn

func Init() {
	connStr := os.Getenv("DATABASE_URL")
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		panic(err)
	}
	if err := conn.Ping(context.Background()); err != nil {
		panic(err)
	}
	Db = conn
}

func CloseDB() error {
	return Db.Close(context.Background())
}

func Migrate() {
	if err := Db.Ping(context.Background()); err != nil {
		panic(err)
	}

	driverURL := "postgres://postgres:abcd1234@localhost:5432/azflowcore?sslmode=disable"
	m, err := migrate.New("file://internal/pkg/db/migrations/postgresql", driverURL)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}
}
