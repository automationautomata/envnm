package main

import (
	"database/sql"
	"envmn/config"
	"flag"
	"log"

	"github.com/pressly/goose/v3"
)

const defaultMigrationsDir = "../../migrations"

func main() {
	connStr := flag.String("conn", "", "connection string")
	migrationsDir := flag.String("conn", defaultMigrationsDir, "connection string")
	flag.Parse()

	if *connStr == "" {
		cfg, err := config.Load[config.PostgresDBConfig]()
		if err != nil {
			log.Fatal(err)
		}
		*connStr = cfg.DSN()
	}

	db, err := sql.Open("postgres", *connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := goose.Up(db, *migrationsDir); err != nil {
		log.Fatal(err)
	}
}
