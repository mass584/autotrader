package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mass584/autotrader/config"
)

func main() {
	mode := os.Args[1]

	if mode != "up" && mode != "down" {
		log.Fatal("Invalid mode")
		return
	}

	config, error := config.NewConfig()

	if error != nil {
		log.Fatal(error)
		return
	}

	db, error := sql.Open("mysql", config.DatabaseURL())

	if error != nil {
		log.Fatal(error)
		return
	}

	driver, error := mysql.WithInstance(db, &mysql.Config{})

	if error != nil {
		log.Fatal(error)
		return
	}

	exec_path, error := os.Getwd()

	if error != nil {
		log.Fatal(error)
		return
	}

	source_url :=
		"file://" +
			exec_path + "/database/migrations"

	migrator, error := migrate.NewWithDatabaseInstance(
		source_url,
		"mysql",
		driver,
	)

	if error != nil {
		log.Fatal(error)
		return
	}

	if mode == "down" {
		error = migrator.Steps(-1)
	} else if mode == "up" {
		error = migrator.Steps(1)
	}

	if error != nil {
		log.Fatal(error)
		return
	}
}
