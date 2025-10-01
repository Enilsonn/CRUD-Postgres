package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
	"github.com/lib/pq"
)

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderWithoutItems = errors.New("order must contain at least one item")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, o *model.Order) (int64, error) {
	if o == nil {
		return 0, errors.New("order payload is nil")
	}
	if len(o.Items) == 0 {
		return 0, ErrOrderWithoutItems
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin order transaction: %w", err)
	}
	defer tx.Rollback()

	var createdAt time.Time
	err = tx.QueryRowContext(ctx, `INSERT INTO orders (client_id, seller_id, payment_method, payment_status)
        VALUES ($1, $2, $3, 'PENDING')
        RETURNING id, created_at`,
		o.ClientID, o.SellerID, o.PaymentMethod,
	).Scan(&o.ID, &createdAt)
	if err != nil {
		return 0, fmt.Errorf("insert order header: %w", err)
	}

	priceStmt, err := tx.PrepareContext(ctx, `SELECT price_cents FROM plans WHERE id = $1`)
	if err != nil {
		return 0, fmt.Errorf("prepare price statement: %w", err)
	}
	defer priceStmt.Close()

	itemStmt, err := tx.PrepareContext(ctx, `INSERT INTO order_items (order_id, plan_id, quantity, unit_price_cents)
        VALUES ($1, $2, $3, $4) RETURNING id`)
	if err != nil {
		return 0, fmt.Errorf("prepare order item insert: %w", err)
	}
	defer itemStmt.Close()

	var subtotal int64
	for idx := range o.Items {
		item := &o.Items[idx]
		if item.Quantity <= 0 {
			return 0, fmt.Errorf("invalid quantity for plan %d", item.PlanID)
		}

		var price int64
		if err := priceStmt.QueryRowContext(ctx, item.PlanID).Scan(&price); err != nil {
			if err == sql.ErrNoRows {
				return 0, fmt.Errorf("plan %d not found", item.PlanID)
			}
			return 0, fmt.Errorf("query plan price for plan %d: %w", item.PlanID, err)
		}

		var itemID int64
		if err := itemStmt.QueryRowContext(ctx, o.ID, item.PlanID, item.Quantity, price).Scan(&itemID); err != nil {
			return 0, fmt.Errorf("insert order item for plan %d: %w", item.PlanID, err)
		}

		item.ID = itemID
		item.OrderID = o.ID
		item.UnitPriceCents = price

		subtotal += price * int64(item.Quantity)
	}

	if _, err := tx.ExecContext(ctx, `UPDATE orders SET subtotal_cents=$1, discount_cents=0, total_cents=$1 WHERE id=$2`, subtotal, o.ID); err != nil {
		return 0, fmt.Errorf("update order totals: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit order: %w", err)
	}

	o.CreatedAt = createdAt
	o.PaymentStatus = "PENDING"
	o.SubtotalCents = subtotal
	o.DiscountCents = 0
	o.TotalCents = subtotal

	return o.ID, nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id int64) (*model.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	order := &model.Order{}
	err := r.db.QueryRowContext(ctx, `SELECT id, client_id, seller_id, created_at, payment_method::text, payment_status::text, subtotal_cents, discount_cents, total_cents
        FROM orders WHERE id = $1`, id).Scan(
		&order.ID,
		&order.ClientID,
		&order.SellerID,
		&order.CreatedAt,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.SubtotalCents,
		&order.DiscountCents,
		&order.TotalCents,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("query order %d: %w", id, err)
	}

	items, err := r.fetchItems(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	order.Items = items[id]

	return order, nil
}

func (r *OrderRepository) ListOrdersByClient(ctx context.Context, clientID int64) ([]model.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `SELECT id, client_id, seller_id, created_at, payment_method::text, payment_status::text, subtotal_cents, discount_cents, total_cents
        FROM orders WHERE client_id = $1 ORDER BY created_at DESC, id DESC`, clientID)
	if err != nil {
		return nil, fmt.Errorf("list orders for client %d: %w", clientID, err)
	}
	defer rows.Close()

	var orders []model.Order
	var orderIDs []int64
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.ID,
			&order.ClientID,
			&order.SellerID,
			&order.CreatedAt,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.SubtotalCents,
			&order.DiscountCents,
			&order.TotalCents,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
		orderIDs = append(orderIDs, order.ID)
	}

	if len(orderIDs) == 0 {
		return orders, nil
	}

	itemsByOrder, err := r.fetchItems(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	for idx := range orders {
		orders[idx].Items = itemsByOrder[orders[idx].ID]
	}

	return orders, nil
}

func (r *OrderRepository) fetchItems(ctx context.Context, orderIDs []int64) (map[int64][]model.OrderItem, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, order_id, plan_id, quantity, unit_price_cents FROM order_items WHERE order_id = ANY($1) ORDER BY id`, pq.Array(orderIDs))
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]model.OrderItem, len(orderIDs))
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.PlanID, &item.Quantity, &item.UnitPriceCents); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		result[item.OrderID] = append(result[item.OrderID], item)
	}

	return result, nil
}

func (r *OrderRepository) FinalizeOrder(ctx context.Context, orderID int64) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := r.db.ExecContext(ctx, `SELECT sp_finalize_order($1)`, orderID); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			lowered := strings.ToLower(pqErr.Message)
			if strings.Contains(lowered, "insufficient stock") {
				return fmt.Errorf("finalize order %d: %w", orderID, ErrInsufficientStock)
			}
		}
		return fmt.Errorf("finalize order %d: %w", orderID, err)
	}

	return nil
}
