package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Enilsonn/CRUD-Postgres/cmd/configs"
	_ "github.com/golang-migrate/migrate/v4"                   // EXECUTA AS MIGRATIONS
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // ADAPTA AS MIGRATIONS POSTGRES
	_ "github.com/golang-migrate/migrate/v4/source/file"       // LER ARQUIVO DAS MIGRATIONS
	"github.com/lib/pq"                                        // DRIVER DO POSTGRES
)

func CreateDatabase(conf configs.DBConf) error {
	postgresDNS := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=desable",
		conf.Host,
		conf.Port,
		conf.User,
		conf.Pass,
		conf.Database,
	)

	conn, err := sql.Open("postgres", postgresDNS)
	if err != nil {
		return fmt.Errorf("error to open connection with the postgres SGDB: %v", err)
	}
	defer conn.Close()

	if err = conn.Ping(); err != nil {
		return fmt.Errorf("error to ping the connection: %v", err)
	}

	// o ideal era colocar essas queries todas na pasta database/queries (fica para a parte 2)
	sql := fmt.Sprintf("CREATE DATABASE %s", conf.Database)
	if _, err = conn.Exec(sql); err != nil {

		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "42P04" {
			log.Printf("Database '%s' alredy exists", conf.Database)
			return nil // nao é um erro: o banco já existe
		}
		return fmt.Errorf("error to create database: %v", err)
	}

	log.Printf("Database '%s' created successfully", conf.Database)

	return nil

}
