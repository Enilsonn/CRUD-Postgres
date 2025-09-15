package database

import (
	"database/sql"
	"fmt"

	"github.com/Enilsonn/CRUD-Postgres/cmd/configs"
	_ "github.com/lib/pq"
)

func OpenConnection() (*sql.DB, error) {
	conf := configs.GetDB()

	sc := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.Host, conf.Port, conf.User, conf.Pass, conf.Database)

	conn, err := sql.Open("postgres", sc)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	err = conn.Ping()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return conn, nil
}
