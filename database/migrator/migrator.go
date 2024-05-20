package main

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mass584/auto-trade/config"
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

	database_url :=
		config.DatabaseUser + ":" + config.DatabasePass +
			"@tcp(" + config.DatabaseHost + ":" + strconv.Itoa(config.DatabasePort) + ")" +
			"/" + config.DatabaseName +
			"?multiStatements=true"

	db, error := sql.Open("mysql", database_url)

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
		error = migrator.Down()
	} else if mode == "up" {
		error = migrator.Up()
	}

	if error != nil {
		log.Fatal(error)
		return
	}
}
