package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
)

var errEmptySellerName = errors.New("seller name cannot be empty")

type SellerRepository struct {
	db *sql.DB
}

func NewSellerRepository(db *sql.DB) *SellerRepository {
	return &SellerRepository{db: db}
}

func (r *SellerRepository) GetOrCreateByName(ctx context.Context, name string) (int64, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return 0, errEmptySellerName
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx for seller get or create: %w", err)
	}
	defer tx.Rollback()

	var id int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM sellers WHERE lower(name) = lower($1) LIMIT 1`, trimmed).Scan(&id)
	if err == nil {
		if err := tx.Commit(); err != nil {
			return 0, fmt.Errorf("commit seller get: %w", err)
		}
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("query seller by name: %w", err)
	}

	if err := tx.QueryRowContext(ctx, `INSERT INTO sellers (name) VALUES ($1) RETURNING id`, trimmed).Scan(&id); err != nil {
		return 0, fmt.Errorf("insert seller: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit seller insert: %w", err)
	}

	return id, nil
}

func (r *SellerRepository) List(ctx context.Context) ([]model.Seller, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM sellers ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list sellers: %w", err)
	}
	defer rows.Close()

	var sellers []model.Seller
	for rows.Next() {
		var seller model.Seller
		if err := rows.Scan(&seller.ID, &seller.Name); err != nil {
			return nil, fmt.Errorf("scan seller: %w", err)
		}
		sellers = append(sellers, seller)
	}

	return sellers, nil
}

func (r *SellerRepository) GetByID(ctx context.Context, id int64) (*model.Seller, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var seller model.Seller
	if err := r.db.QueryRowContext(ctx, `SELECT id, name FROM sellers WHERE id = $1`, id).Scan(&seller.ID, &seller.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("seller %d not found", id)
		}
		return nil, fmt.Errorf("query seller %d: %w", id, err)
	}

	return &seller, nil
}
