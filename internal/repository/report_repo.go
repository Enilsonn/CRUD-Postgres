package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
)

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) ListSellerMonthlySales(ctx context.Context, month *time.Time) ([]model.SellerMonthlySales, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var (
		rows *sql.Rows
		err  error
	)

	if month != nil {
		rows, err = r.db.QueryContext(ctx, `SELECT month, seller_id, orders_count, total_cents
            FROM seller_monthly_sales
            WHERE month = date_trunc('month', $1::timestamptz)
            ORDER BY seller_id`, month.UTC())
	} else {
		rows, err = r.db.QueryContext(ctx, `SELECT month, seller_id, orders_count, total_cents
            FROM seller_monthly_sales
            ORDER BY month DESC, seller_id`)
	}
	if err != nil {
		return nil, fmt.Errorf("query seller monthly sales: %w", err)
	}
	defer rows.Close()

	var results []model.SellerMonthlySales
	for rows.Next() {
		var record model.SellerMonthlySales
		if err := rows.Scan(&record.Month, &record.SellerID, &record.OrdersCount, &record.TotalCents); err != nil {
			return nil, fmt.Errorf("scan seller monthly sales: %w", err)
		}
		results = append(results, record)
	}

	return results, nil
}
