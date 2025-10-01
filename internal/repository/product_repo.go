package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

type PlanSearchFilters struct {
	Name               string
	MinPriceCents      *int64
	MaxPriceCents      *int64
	Category           string
	ManufacturedInMari *bool
}

func (r *ProductRepository) CreateClientProduct(ctx context.Context, plan model.Plan) (int64, error) {
	sql := `INSERT INTO plans (plan_name, price_cents, amount_credits, status, category, manufactured_in_mari, stock)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
	`
	var id int64
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.db.QueryRowContext(
		ctx,
		sql,
		plan.PlanName,
		plan.PriceCents,
		plan.AmountCredits,
		plan.Status,
		plan.Category,
		plan.ManufacturedInMari,
		plan.Stock,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error to create a new plan: %w", err)
	}

	return id, nil
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id int64) (*model.Plan, error) {
	sql := `SELECT id, plan_name, price_cents, amount_credits, status, category, manufactured_in_mari, stock
			FROM plans
			WHERE id=$1
			AND status=true
	`
	var plan model.Plan
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(
		ctx,
		sql,
		id,
	)
	err := row.Scan(
		&plan.ID,
		&plan.PlanName,
		&plan.PriceCents,
		&plan.AmountCredits,
		&plan.Status,
		&plan.Category,
		&plan.ManufacturedInMari,
		&plan.Stock,
	)
	if err != nil {
		return nil, fmt.Errorf("error to find plan with id %d: %w", id, err)
	}

	return &plan, nil
}

func (r *ProductRepository) GetClientProductByName(ctx context.Context, plan_name string) (*model.Plan, error) {
	sql := `SELECT id, plan_name, price_cents, amount_credits, status, category, manufactured_in_mari, stock
			FROM plans
			WHERE plan_name=$1
	`
	var plan model.Plan
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(
		ctx,
		sql,
		plan_name,
	)
	err := row.Scan(
		&plan.ID,
		&plan.PlanName,
		&plan.PriceCents,
		&plan.AmountCredits,
		&plan.Status,
		&plan.Category,
		&plan.ManufacturedInMari,
		&plan.Stock,
	)
	if err != nil {
		return nil, fmt.Errorf("error to find plan with name %s: %w", plan_name, err)
	}

	return &plan, nil
}

func (r *ProductRepository) GetAllClientProduct(ctx context.Context) ([]model.Plan, error) {
	sql := `SELECT id, plan_name, price_cents, amount_credits, status, category, manufactured_in_mari, stock
			FROM plans
			WHERE status=true
	`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(
		ctx,
		sql,
	)
	if err != nil {
		return nil, fmt.Errorf("error to find plans list: %w", err)
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
			&plan.Category,
			&plan.ManufacturedInMari,
			&plan.Stock,
		)
		if err != nil {
			return nil, fmt.Errorf("error to access plan: %w", err)
		}

		plans = append(plans, plan)
	}

	return plans, nil
}

func (r *ProductRepository) UpdateClientProduct(ctx context.Context, id int64, plan model.Plan) (int64, error) {
	sql := `UPDATE plans
			SET plan_name=$1, price_cents=$2, amount_credits=$3, category=$4, manufactured_in_mari=$5, stock=$6
			WHERE id=$7
			AND status=true
	`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := r.db.ExecContext(
		ctx,
		sql,
		plan.PlanName,
		plan.PriceCents,
		plan.AmountCredits,
		plan.Category,
		plan.ManufacturedInMari,
		plan.Stock,
		id,
	)
	if err != nil {
		return 0, fmt.Errorf("error to update plan %d: %w", id, err)
	}

	rowsAffected, err := resp.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("unable to check rows affected on update: %w", err)
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("no plan found with id %d was updated", id)
	}

	return rowsAffected, nil
}

func (r *ProductRepository) DeleteClientProduct(ctx context.Context, id int64) error {
	sql := `UPDATE plans
			SET status=false
			WHERE id=$1
	`
	resp, err := r.db.ExecContext(
		ctx,
		sql,
		id,
	)
	if err != nil {
		return fmt.Errorf("error to delete plan %d: %w", id, err)
	}

	rowsAffected, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to check rows affected on delete: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no plan found with id %d was updated", id)
	}

	return nil
}

func (r *ProductRepository) SearchPlans(ctx context.Context, filters PlanSearchFilters) ([]model.Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := strings.Builder{}
	query.WriteString(`SELECT id, plan_name, price_cents, amount_credits, status, category, manufactured_in_mari, stock FROM plans WHERE status = true`)

	var args []any
	argPos := 1

	if filters.Name != "" {
		query.WriteString(fmt.Sprintf(" AND plan_name ILIKE $%d", argPos))
		args = append(args, "%"+filters.Name+"%")
		argPos++
	}
	if filters.MinPriceCents != nil {
		query.WriteString(fmt.Sprintf(" AND price_cents >= $%d", argPos))
		args = append(args, *filters.MinPriceCents)
		argPos++
	}
	if filters.MaxPriceCents != nil {
		query.WriteString(fmt.Sprintf(" AND price_cents <= $%d", argPos))
		args = append(args, *filters.MaxPriceCents)
		argPos++
	}
	if filters.Category != "" {
		query.WriteString(fmt.Sprintf(" AND category = $%d", argPos))
		args = append(args, filters.Category)
		argPos++
	}
	if filters.ManufacturedInMari != nil {
		query.WriteString(fmt.Sprintf(" AND manufactured_in_mari = $%d", argPos))
		args = append(args, *filters.ManufacturedInMari)
		argPos++
	}

	query.WriteString(" ORDER BY plan_name")

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("error searching plans: %w", err)
	}
	defer rows.Close()

	var plans []model.Plan
	for rows.Next() {
		var plan model.Plan
		if err := rows.Scan(
			&plan.ID,
			&plan.PlanName,
			&plan.PriceCents,
			&plan.AmountCredits,
			&plan.Status,
			&plan.Category,
			&plan.ManufacturedInMari,
			&plan.Stock,
		); err != nil {
			return nil, fmt.Errorf("error scanning plan: %w", err)
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

func (r *ProductRepository) ListLowStock(ctx context.Context) ([]model.Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `SELECT id, plan_name, price_cents, amount_credits, status, category, manufactured_in_mari, stock FROM plans WHERE status = true AND stock < 5 ORDER BY stock ASC, plan_name`)
	if err != nil {
		return nil, fmt.Errorf("error listing low stock plans: %w", err)
	}
	defer rows.Close()

	var plans []model.Plan
	for rows.Next() {
		var plan model.Plan
		if err := rows.Scan(
			&plan.ID,
			&plan.PlanName,
			&plan.PriceCents,
			&plan.AmountCredits,
			&plan.Status,
			&plan.Category,
			&plan.ManufacturedInMari,
			&plan.Stock,
		); err != nil {
			return nil, fmt.Errorf("error scanning plan: %w", err)
		}
		plans = append(plans, plan)
	}

	return plans, nil
}
