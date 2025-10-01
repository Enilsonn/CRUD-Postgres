package repository

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
)

type PricingRepository struct {
	db *sql.DB
}

func NewPricingRepository(db *sql.DB) *PricingRepository {
	return &PricingRepository{db: db}
}

func (r *PricingRepository) FindRate(ctx context.Context, model string) (ppk, cpk float64, ok bool, err error) {
	const query = `
		SELECT pattern, credits_per_1k_prompt, credits_per_1k_completion
		FROM model_pricing
		WHERE active = true
		ORDER BY priority ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return 0, 0, false, fmt.Errorf("query model pricing: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			pattern        string
			promptRate     float64
			completionRate float64
		)

		if err := rows.Scan(&pattern, &promptRate, &completionRate); err != nil {
			return 0, 0, false, fmt.Errorf("scan pricing row: %w", err)
		}

		re, err := regexp.Compile(pattern)
		if err != nil {
			return 0, 0, false, fmt.Errorf("invalid pricing pattern %q: %w", pattern, err)
		}

		if re.MatchString(model) {
			return promptRate, completionRate, true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return 0, 0, false, fmt.Errorf("iterate pricing rows: %w", err)
	}

	return 0, 0, false, nil
}

func (r *PricingRepository) ListActive(ctx context.Context) ([]*model.ModelPricing, error) {
	const query = `
		SELECT id, pattern, credits_per_1k_prompt, credits_per_1k_completion, priority, active, updated_at
		FROM model_pricing
		WHERE active = true
		ORDER BY priority ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query active pricing: %w", err)
	}
	defer rows.Close()

	var rates []*model.ModelPricing
	for rows.Next() {
		rate := &model.ModelPricing{}
		if err := rows.Scan(
			&rate.ID,
			&rate.Pattern,
			&rate.CreditsPer1KPrompt,
			&rate.CreditsPer1KCompletion,
			&rate.Priority,
			&rate.Active,
			&rate.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pricing row: %w", err)
		}
		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pricing rows: %w", err)
	}

	return rates, nil
}

func (r *PricingRepository) Upsert(ctx context.Context, rate *model.ModelPricing) (*model.ModelPricing, error) {
	if rate == nil {
		return nil, fmt.Errorf("rate payload is required")
	}

	const returning = ` RETURNING id, pattern, credits_per_1k_prompt, credits_per_1k_completion, priority, active, updated_at`

	var row *sql.Row
	if rate.ID > 0 {
		row = r.db.QueryRowContext(
			ctx,
			`UPDATE model_pricing
			SET pattern = $1,
				credits_per_1k_prompt = $2,
				credits_per_1k_completion = $3,
				priority = $4,
				active = $5,
				updated_at = NOW()
			WHERE id = $6`+returning,
			rate.Pattern,
			rate.CreditsPer1KPrompt,
			rate.CreditsPer1KCompletion,
			rate.Priority,
			rate.Active,
			rate.ID,
		)
	} else {
		row = r.db.QueryRowContext(
			ctx,
			`INSERT INTO model_pricing (pattern, credits_per_1k_prompt, credits_per_1k_completion, priority, active)
			VALUES ($1, $2, $3, $4, $5)`+returning,
			rate.Pattern,
			rate.CreditsPer1KPrompt,
			rate.CreditsPer1KCompletion,
			rate.Priority,
			rate.Active,
		)
	}

	updated := &model.ModelPricing{}
	if err := row.Scan(
		&updated.ID,
		&updated.Pattern,
		&updated.CreditsPer1KPrompt,
		&updated.CreditsPer1KCompletion,
		&updated.Priority,
		&updated.Active,
		&updated.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("persist pricing rate: %w", err)
	}

	return updated, nil
}
