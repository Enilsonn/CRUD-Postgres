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

func (r *ProductRepository) CreateClientProduct(plan model.Plan) (int64, error) {
	sql := `INSERT INTO plans (plan_name, price_cents, amount_credits, status)
			VALUES ($1, $2, $3, $4)
			RETURNING id
	`
	var id int64
	err := r.db.QueryRow(
		sql,
		plan.PlanName,
		plan.PriceCents,
		plan.AmountCredits,
		plan.Status,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error to create a new plan: %v", err)
	}

	return id, nil
}

func (r *ProductRepository) GetProductByID(id int64) (*model.Plan, error) {
	sql := `SELECT id, plan_name, price_cents, amount_credits, status
			FROM plans
			WHERE id=$1
	`
	var plan model.Plan
	row := r.db.QueryRow(
		sql,
		id,
	)
	err := row.Scan(
		&plan.ID,
		&plan.PlanName,
		&plan.PriceCents,
		&plan.AmountCredits,
		&plan.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("error to found plan with id %d: %v", id, err)
	}

	return &plan, nil
}

func (r *ProductRepository) GetClientProductByName(plan_name string) (*model.Plan, error) {
	sql := `SELECT id, plan_name, price_cents, amount_credits, status
			FROM plans
			WHERE plan_name=$1
	`
	var plan model.Plan
	row := r.db.QueryRow(
		sql,
		plan_name,
	)
	err := row.Scan(
		&plan.ID,
		&plan.PlanName,
		&plan.PriceCents,
		&plan.AmountCredits,
		&plan.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("error to found plan with name %s: %v", plan_name, err)
	}

	return &plan, nil
}

func (r *ProductRepository) GetAllClientProduct() ([]model.Plan, error) {
	sql := `SELECT id, plan_name, price_cents, amount_credits, status
			FROM plans
			WHERE status=true
	`
	rows, err := r.db.Query(
		sql,
	)
	if err != nil {
		return nil, fmt.Errorf("error to found plans list: %v", err)
	}
	defer rows.Close()

	var plans []model.Plan
	for rows.Next() {
		var plan model.Plan

		err := rows.Scan(
			&plan.ID,
			&plan.PlanName,
			&plan.PriceCents,
			&plan.AmountCredits,
			&plan.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("error to access plan: %v", err)
		}

		plans = append(plans, plan)
	}

	return plans, nil
}

func (r *ProductRepository) UpdateClientProduct(id int64, plan model.Plan) (int64, error) {
	sql := `UPDATE plans
			SET plan_name=$1, price_cents=$2, amount_credits=$3
			WHERE id=$4
	`

	resp, err := r.db.Exec(
		sql,
		plan.PlanName,
		plan.PriceCents,
		plan.AmountCredits,
		id,
	)
	if err != nil {
		return 0, fmt.Errorf("error to update plan %d: %v", id, err)
	}

	rowsAffected, err := resp.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("not possivel check rows affected on update: %v", err)
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("no plan found with id %d was updated", id)
	}

	return rowsAffected, nil
}

func (r *ProductRepository) DeleteClientProduct(id int64) error {
	sql := `UPDATE plans
			SET status=false
			WHERE id=$1
	`
	resp, err := r.db.Exec(
		sql,
		id,
	)
	if err != nil {
		return fmt.Errorf("error to delete plan %d: %v", id, err)
	}

	rowsAffected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("not possivel check rows affected on delete: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no plan found with id %d was updated", id)
	}

	return nil
}
