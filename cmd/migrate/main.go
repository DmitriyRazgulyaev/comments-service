package main

import (
	"comments-service/internal/config"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var migrationsPath string
	var direction string
	var version int

	flag.StringVar(&migrationsPath, "path", "./migrations", "path to migrations directory")
	flag.StringVar(&direction, "direction", "up", "migration direction: up, down or force")
	flag.IntVar(&version, "version", -1, "migration version for force")
	flag.Parse()

	envPath := os.Getenv("ENV_PATH")
	if envPath == "" {
		envPath = "./config/.env"
	}

	cfg, err := config.ParseConfigFromEnv(envPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	dsn := &url.URL{
		Scheme: "postgres",
		User: url.UserPassword(
			cfg.DBConfig.User,
			cfg.DBConfig.Password,
		),
		Host: fmt.Sprintf(
			"%s:%v",
			cfg.DBConfig.Host,
			cfg.DBConfig.Port,
		),
		Path: "/" + cfg.DBConfig.DBName,
	}

	query := dsn.Query()
	query.Set("sslmode", "disable")
	dsn.RawQuery = query.Encode()

	m, err := migrate.New(fmt.Sprintf("file://%s", migrationsPath), dsn.String())
	if err != nil {
		log.Fatalf("initialize migrations: %v", err)
	}

	defer func() {
		sourceErr, databaseErr := m.Close()
		if sourceErr != nil {
			log.Printf("close migration source: %v", sourceErr)
		}
		if databaseErr != nil {
			log.Printf("close migration database: %v", databaseErr)
		}
	}()

	switch direction {
	case "up":
		err = m.Up()

	case "down":
		err = m.Steps(-1)

	case "force":
		if version < 0 {
			log.Fatal("version must be specified for force")
		}

		err = m.Force(version)

	default:
		log.Fatalf("unknown direction: %s", direction)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("No migrations to apply")
		return
	}

	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Printf("Migrations %s completed successfully", direction)
}
