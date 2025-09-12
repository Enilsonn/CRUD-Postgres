package repository

import (
	"database/sql"
	"fmt"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
)

type ClientRepository struct {
	db *sql.DB
}

func NewClientRepository(db *sql.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) CreateClient(client model.Client) (int64, error) {
	sql := `INSERT INTO clients (name, email, phone, status,registration_data)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
	`

	var id int64

	err := r.db.QueryRow(
		sql,
		client.Name,
		client.Email,
		client.Phone,
		client.Status,
		client.RegistrationData,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error creating client on the table: %v", err)
	}

	return id, nil
}

func (r *ClientRepository) GetClientByID(id int64) (*model.Client, error) {
	sql := `SELECT id, name, email, phone, status, registration_data
			FROM clients
			WHERE id=$1
	`

	var client model.Client

	row := r.db.QueryRow(
		sql,
		id,
	)
	err := row.Scan(
		&client.ID,
		&client.Name,
		&client.Email,
		&client.Phone,
		&client.Status,
		&client.RegistrationData,
	)
	if err != nil {
		return nil, fmt.Errorf("error to found client %d: %v", id, err)
	}

	return &client, nil
}

func (r *ClientRepository) GetAllClients() ([]model.Client, error) {
	sql := `SELECT id, name, email, phone, status, registration_data
			FROM clients
			WHERE status=true
	`

	rows, err := r.db.Query(
		sql,
	)
	if err != nil {
		return nil, fmt.Errorf("error to found client list: %v", err)
	}
	defer rows.Close()

	var clients []model.Client
	for rows.Next() {
		var client model.Client

		err := rows.Scan(
			&client.ID,
			&client.Name,
			&client.Email,
			&client.Phone,
			&client.Status,
			&client.RegistrationData,
		)
		if err != nil {
			return nil, fmt.Errorf("error to access client: %v", err)
		}

		clients = append(clients, client)
	}

	return clients, nil
}

func (r *ClientRepository) GetClientByName(name string) ([]model.Client, error) {
	sql := `SELECT id, name, email, phone, status, registration_data
			FROM clients
			WHERE name=$1
	`

	rows, err := r.db.Query(sql, name)
	if err != nil {
		return nil, fmt.Errorf("error to found clients named %s: %v", name, err)
	}
	defer rows.Close()

	var clients []model.Client
	for rows.Next() {
		var client model.Client

		err := rows.Scan(
			&client.ID,
			&client.Name,
			&client.Email,
			&client.Phone,
			&client.Status,
			&client.RegistrationData,
		)
		if err != nil {
			return nil, fmt.Errorf("error to access client: %v", err)
		}

		clients = append(clients, client)
	}

	return clients, nil
}

func (r *ClientRepository) UpdateClients(id int64, client model.Client) (int64, error) {
	sql := `UPDATE clients
			SET name=$1, email=$2, phone=$3
			WHERE id=$4
	`
	resp, err := r.db.Exec(
		sql,
		client.Name,
		client.Email,
		client.Phone,
		id,
	)
	if err != nil {
		return 0, fmt.Errorf("error to update client %d: %v", id, err)
	}

	rowsAffected, err := resp.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("not possivel check rows affected on update: %v", err)
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("no client found with id %d was updated", id)
	}

	return rowsAffected, nil
}

func (r *ClientRepository) DeleteClient(id int64) error {
	sql := `UPDATE clients
			SET status=false
			WHERE id=$1
	`

	resp, err := r.db.Exec(
		sql,
		id,
	)
	if err != nil {
		return fmt.Errorf("error to delete client %d: %v", id, err)
	}

	rowsAffected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("error to check rows affected on delete: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no client found with id %d was deleted", id)
	}

	return nil
}
