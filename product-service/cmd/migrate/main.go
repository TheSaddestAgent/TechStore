package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"techstore/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	var migrationsPath string
	var command string

	flag.StringVar(&migrationsPath, "path", "./migrations", "path to migrations directory")
	flag.StringVar(&command, "command", "up", "migration command: up, down, force, version")
	flag.Parse()

	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: could not load .env file: %v", err)
	}

	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("error parsing config: %v", err)
	}

	// Формируем DSN для подключения к PostgreSQL
	hostPort := net.JoinHostPort(cfg.Host, cfg.Port)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.Username,
		cfg.Password,
		hostPort,
		cfg.DBName,
	)

	log.Printf("Connecting to database: %s@%s", cfg.Username, hostPort)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dsn,
	)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}
	defer m.Close()

	switch command {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("failed to apply migrations: %v", err)
		}
		log.Println("Migrations applied successfully!")

	case "down":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("failed to rollback migrations: %v", err)
		}
		log.Println("Migrations rolled back successfully!")

	case "force":
		version, dirty, errVersion := m.Version()
		if errVersion != nil {
			fmt.Printf("failed to get version: %v\n", errVersion)
			os.Exit(1)
		}
		fmt.Printf("Current version: %d, Dirty: %v\n", version, dirty)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("failed to get version: %v", err)
		}
		log.Printf("Current version: %d, Dirty: %v", version, dirty)

	default:
		log.Printf("unknown command: %s", command)
		log.Println("Available commands: up, down, force, version")
		os.Exit(1)
	}
}
