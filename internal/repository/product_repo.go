package repository

import (
	"database/sql"
	"fmt"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) CreateClientProduct(client_product model.ClientProduct) (int64, error) {
	sql := `INSERT INTO client_product (id, plan_name, price_cents, amount_credits, status)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
	`
	var id int64
	err := r.db.QueryRow(
		sql,
		client_product.ID, // chave estrangeira
		client_product.PlanName,
		client_product.PriceCents,
		client_product.AmountCredits,
		client_product.Status,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error to create a new client-procuct: %v", err)
	}

	return id, nil
}

func (r *ProductRepository) GetProductByID(id int64) (*model.ClientProduct, error) {
	sql := `SELECT id, plan_name, price_cents, amount_credits, status
			FROM client_product
			WHERE id=$1
			AND status=true
	`
	var client_product model.ClientProduct
	row := r.db.QueryRow(
		sql,
		id,
	)
	err := row.Scan(
		&client_product.ID,
		&client_product.PlanName,
		&client_product.PriceCents,
		&client_product.AmountCredits,
		&client_product.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("error to found client-product with id %d: %v", id, err)
	}

	return &client_product, nil
}

func (r *ProductRepository) GetClientProductByName(plan_name string) (*model.ClientProduct, error) {
	sql := `SELECT id, plan_name, price_cents, amount_credits, status
			FROM client_product
			WHERE name=$1
			AND status=true
	`
	var client_product model.ClientProduct
	row := r.db.QueryRow(
		sql,
		plan_name,
	)
	err := row.Scan(
		&client_product.ID,
		&client_product.PlanName,
		&client_product.PriceCents,
		&client_product.AmountCredits,
		&client_product.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("error to found client-product with name %s: %v", plan_name, err)
	}

	return &client_product, nil
}

func (r *ProductRepository) GetAllClientProduct() ([]model.ClientProduct, error) {
	sql := `SELECT id, plan_name, price_cents, amount_credits, status
			FROM client_product
			WHERE status=true
	`
	rows, err := r.db.Query(
		sql,
	)
	if err != nil {
		return nil, fmt.Errorf("error to found client-product list: %v", err)
	}
	defer rows.Close()

	var client_products []model.ClientProduct
	for rows.Next() {
		var client_product model.ClientProduct

		err := rows.Scan(
			&client_product.ID,
			&client_product.PlanName,
			&client_product.PriceCents,
			&client_product.AmountCredits,
			&client_product.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("error to access client-product: %v", err)
		}

		client_products = append(client_products, client_product)
	}

	return client_products, nil
}

func (r *ProductRepository) UpdateClientProduct(id int64, client_product model.ClientProduct) (int64, error) {
	sql := `UPDATE client_product
			SET plan_name=$1, price_cents=$2, amount_credits=$3
			WHERE id=$4
			AND status=true
	`

	resp, err := r.db.Exec(
		sql,
		client_product.PlanName,
		client_product.PriceCents,
		client_product.AmountCredits,
		id,
	)
	if err != nil {
		return 0, fmt.Errorf("error to update client-product %d: %v", id, err)
	}

	rowsAffected, err := resp.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("not possivel check rows affected on update: %v", err)
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("no client-product found with id %d was updated", id)
	}

	return rowsAffected, nil
}

func (r *ProductRepository) DeleteClientProduct(id int64) error {
	sql := `UPDATE client_product
			SET status=false
			WHERE id=$1
	`
	resp, err := r.db.Exec(
		sql,
		id,
	)
	if err != nil {
		return fmt.Errorf("error to delete client-product %d: %v", id, err)
	}

	rowsAffected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("not possivel check rows affected on delete: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no client-product found with id %d was updated", id)
	}

	return nil
}
