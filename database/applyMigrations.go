package database

import (
	"fmt"
	"log"

	"github.com/Enilsonn/CRUD-Postgres/cmd/configs"
	"github.com/golang-migrate/migrate/v4"                     // EXECUTA AS MIGRATIONS
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // ADAPTA AS MIGRATIONS POSTGRES
	_ "github.com/golang-migrate/migrate/v4/source/file"       // LER ARQUIVO DAS MIGRATIONS
	_ "github.com/lib/pq"                                      // DRIVER DO POSTGRES
)

func ApplyMigrations(conf configs.DBConf) error {
	databaseDNS := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.User,
		conf.Pass,
		conf.Host,
		conf.Port,
		conf.Database,
	)

	m, err := migrate.New("file://./migrations", databaseDNS)
	if err != nil {
		return fmt.Errorf("error to inicialize migrations intances: %v", err)
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error to apply migrations: %v", err)
	}

	log.Println("migrations applied successfully")

	return nil
}
