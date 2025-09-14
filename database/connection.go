package database

import (
	"database/sql"
	"fmt"

	"github.com/Enilsonn/CRUD-Postgres/cmd/configs"
	_ "github.com/lib/pq"
)

func OpenConecction(conf configs.DBConf) (*sql.DB, error) {

	sc := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=desable",
		conf.Host, conf.Port, conf.User, conf.Pass, conf.Database)

	conn, err := sql.Open("postgres", sc)
	if err != nil {
		panic(err)
	}

	err = conn.Ping()

	return conn, err
}
