package db

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"log"
	"os"
)

var Conn *pgx.Conn

func Init() {
	connStr := os.Getenv("DATABASE_URL")
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		panic(err)
	}
	if err := conn.Ping(context.Background()); err != nil {
		panic(err)
	}
	Conn = conn
}

func CloseDB() error {
	return Conn.Close(context.Background())
}

func Migrate() {
	driverURL := "postgres://postgres:abcd1234@azflow-db:5432/azflowcore?sslmode=disable"
	m, err := migrate.New("file://db/migration", driverURL)
	if err != nil {
		panic(err)
	}

	// Force the migration version and mark as clean
	// enable this if you want to force the migration
	//err = m.Force(1)
	//if err != nil {
	//	log.Fatalf("failed to force migration version: %v", err)
	//}
	// or run this cli: migrate -database "postgres://postgres:abcd1234@localhost:5432/azflowcore?sslmode=disable" -path ./db/migration force 1

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Printf("migration failed: %v", err)
		if rollbackErr := m.Down(); rollbackErr != nil {
			log.Fatalf("failed to rollback migration: %v", rollbackErr)
		}
		log.Fatalf("migration rolled back due to failure")
	}
}
