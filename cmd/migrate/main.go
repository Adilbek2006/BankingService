package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	var host, port, user, password, dbname, sslmode string
	flag.StringVar(&host, "db-host", "127.0.0.1", "Postgres host")
	flag.StringVar(&port, "db-port", "5433", "Postgres port")
	flag.StringVar(&user, "db-user", "user", "Postgres user")
	flag.StringVar(&password, "db-password", "password", "Postgres password")
	flag.StringVar(&dbname, "db-name", "banking", "Postgres database")
	flag.StringVar(&sslmode, "db-sslmode", "disable", "Postgres sslmode")
	flag.Parse()

	sourceURL, err := migrationsURL()
	if err != nil {
		log.Fatalf("Failed to build migrations path: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("DB open error: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := db.Ping(); err != nil {
		log.Fatalf("DB ping error: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("DB driver error: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		log.Fatalf("Migrate init error: %v", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("Migrations: no changes")
			return
		}
		log.Fatalf("Migrate up error: %v", err)
	}

	log.Println("Migrations applied successfully")
}

func migrationsURL() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	abs, err := filepath.Abs(filepath.Join(wd, "migrations"))
	if err != nil {
		return "", err
	}

	slashPath := strings.ReplaceAll(abs, "\\", "/")
	return "file://" + slashPath, nil
}
