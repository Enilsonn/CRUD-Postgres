package repository

import (
	"context"
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

func (r *ClientRepository) CreateClient(ctx context.Context, client model.Client) (int64, error) {
	sql := `INSERT INTO clients (name, email, phone, status, registration_data, supports_flamengo, watches_one_piece, city)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id
	`

	var id int64

	err := r.db.QueryRowContext(
		ctx,
		sql,
		client.Name,
		client.Email,
		client.Phone,
		client.Status,
		client.RegistrationData,
		client.SupportsFlamengo,
		client.WatchesOnePiece,
		client.City,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error creating client in the table: %w", err)
	}

	return id, nil
}

func (r *ClientRepository) GetClientByID(ctx context.Context, id int64) (*model.Client, error) {
	sql := `SELECT id, name, email, phone, status, registration_data, supports_flamengo, watches_one_piece, city
			FROM clients
			WHERE id=$1
			AND status=true
	`

	var client model.Client

	row := r.db.QueryRowContext(
		ctx,
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
		&client.SupportsFlamengo,
		&client.WatchesOnePiece,
		&client.City,
	)
	if err != nil {
		return nil, fmt.Errorf("error to find client %d: %w", id, err)
	}

	return &client, nil
}

func (r *ClientRepository) GetAllClients(ctx context.Context) ([]model.Client, error) {
	sql := `SELECT id, name, email, phone, status, registration_data, supports_flamengo, watches_one_piece, city
			FROM clients
			WHERE status=true
	`

	rows, err := r.db.QueryContext(
		ctx,
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
			&client.SupportsFlamengo,
			&client.WatchesOnePiece,
			&client.City,
		)
		if err != nil {
			return nil, fmt.Errorf("error to access client: %w", err)
		}

		clients = append(clients, client)
	}

	return clients, nil
}

func (r *ClientRepository) GetClientByName(ctx context.Context, name string) ([]model.Client, error) {
	sql := `SELECT id, name, email, phone, status, registration_data, supports_flamengo, watches_one_piece, city
			FROM clients
			WHERE name=$1
			AND status=true
	`

	rows, err := r.db.QueryContext(
		ctx,
		sql,
		name)
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
			&client.SupportsFlamengo,
			&client.WatchesOnePiece,
			&client.City,
		)
		if err != nil {
			return nil, fmt.Errorf("error to access client: %w", err)
		}

		clients = append(clients, client)
	}

	return clients, nil
}

func (r *ClientRepository) UpdateClients(ctx context.Context, id int64, client model.Client) (int64, error) {
	sql := `UPDATE clients
			SET name=$1, email=$2, phone=$3, supports_flamengo=$4, watches_one_piece=$5, city=$6
			WHERE id=$7
			AND status=true
	`
	resp, err := r.db.ExecContext(
		ctx,
		sql,
		client.Name,
		client.Email,
		client.Phone,
		client.SupportsFlamengo,
		client.WatchesOnePiece,
		client.City,
		id,
	)
	if err != nil {
		return 0, fmt.Errorf("error to update client %d: %w", id, err)
	}

	rowsAffected, err := resp.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("unable to check rows affected on update: %w", err)
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("no client found with id %d was updated", id)
	}

	return rowsAffected, nil
}

func (r *ClientRepository) DeleteClient(ctx context.Context, id int64) error {
	sql := `UPDATE clients
			SET status=false
			WHERE id=$1
	`

	resp, err := r.db.ExecContext(
		ctx,
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
