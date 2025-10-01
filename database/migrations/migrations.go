package migrations

import (
	"database/sql"
	"fmt"
)

func Up(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS clients (
			id               BIGSERIAL PRIMARY KEY,
			name             TEXT        NOT NULL,
			email            TEXT        NOT NULL UNIQUE,
			phone            TEXT        NOT NULL,
			status           BOOLEAN     NOT NULL DEFAULT TRUE,
			registration_data TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS plans (
			id              BIGSERIAL PRIMARY KEY,
			plan_name       TEXT        NOT NULL UNIQUE,
			price_cents     BIGINT      NOT NULL CHECK (price_cents >= 0),
			amount_credits  INT         NOT NULL CHECK (amount_credits > 0),
			status          BOOLEAN     NOT NULL DEFAULT TRUE
		);`,

		`CREATE TABLE IF NOT EXISTS wallets (
			client_id       BIGINT      PRIMARY KEY REFERENCES clients(id) ON DELETE CASCADE,
			balance_credits BIGINT      NOT NULL DEFAULT 0 CHECK (balance_credits >= 0)
		);`,

		`DO $$ BEGIN
			CREATE TYPE credit_type AS ENUM ('TOPUP','USAGE','REFUND','ADJUST');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		`CREATE TABLE IF NOT EXISTS credit_ledger (
			id                BIGSERIAL PRIMARY KEY,
			client_id         BIGINT      NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
			type              credit_type NOT NULL,
			credits_delta     BIGINT      NOT NULL,
			price_cents_delta BIGINT      NOT NULL DEFAULT 0,
			meta              JSONB       NOT NULL DEFAULT '{}',
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS usage_events (
			id               BIGSERIAL PRIMARY KEY,
			client_id        BIGINT      NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
			model            TEXT        NOT NULL,
			prompt_tokens    BIGINT      NOT NULL DEFAULT 0,
			completion_tokens BIGINT     NOT NULL DEFAULT 0,
			credits_spent    BIGINT      NOT NULL DEFAULT 0,
			created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
	}

	for i, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration %d: %w", i+1, err)
		}
	}

	if err := runPhaseTwo(db); err != nil {
		return fmt.Errorf("failed to execute phase 2 migration: %w", err)
	}

	return nil
}

func runPhaseTwo(db *sql.DB) error {
	statements := []string{
		`DO $$ BEGIN
	  CREATE TYPE payment_method AS ENUM ('CARD','BOLETO','PIX','BERRIES');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		`DO $$ BEGIN
	  CREATE TYPE payment_status AS ENUM ('PENDING','CONFIRMED','FAILED','CANCELED');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		`ALTER TABLE clients
  ADD COLUMN IF NOT EXISTS supports_flamengo boolean DEFAULT false,
  ADD COLUMN IF NOT EXISTS watches_one_piece boolean DEFAULT false,
  ADD COLUMN IF NOT EXISTS city text;`,
		`CREATE TABLE IF NOT EXISTS sellers (
  id BIGSERIAL PRIMARY KEY,
  name text NOT NULL
);`,
		`ALTER TABLE plans
  ADD COLUMN IF NOT EXISTS category text NOT NULL DEFAULT 'CREDITS',
  ADD COLUMN IF NOT EXISTS manufactured_in_mari boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS stock int NOT NULL DEFAULT 999999 CHECK (stock >= 0);`,
		`CREATE TABLE IF NOT EXISTS orders (
  id BIGSERIAL PRIMARY KEY,
  client_id bigint NOT NULL REFERENCES clients(id),
  seller_id bigint NOT NULL REFERENCES sellers(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  payment_method payment_method NOT NULL,
  payment_status payment_status NOT NULL DEFAULT 'PENDING',
  subtotal_cents bigint NOT NULL DEFAULT 0 CHECK (subtotal_cents >= 0),
  discount_cents bigint NOT NULL DEFAULT 0 CHECK (discount_cents >= 0),
  total_cents bigint NOT NULL DEFAULT 0 CHECK (total_cents >= 0)
);`,
		`CREATE TABLE IF NOT EXISTS order_items (
  id BIGSERIAL PRIMARY KEY,
  order_id bigint NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  plan_id bigint NOT NULL REFERENCES plans(id),
  quantity int NOT NULL CHECK (quantity > 0),
  unit_price_cents bigint NOT NULL CHECK (unit_price_cents >= 0)
);`,
		`CREATE INDEX IF NOT EXISTS idx_plans_name ON plans USING gin (to_tsvector('simple', plan_name));`,
		`CREATE INDEX IF NOT EXISTS idx_plans_category ON plans(category);`,
		`CREATE INDEX IF NOT EXISTS idx_plans_price ON plans(price_cents);`,
		`CREATE INDEX IF NOT EXISTS idx_plans_mari ON plans(manufactured_in_mari);`,
		`CREATE INDEX IF NOT EXISTS idx_plans_lowstock ON plans(stock) WHERE stock < 5;`,
		`CREATE OR REPLACE VIEW seller_monthly_sales AS
SELECT date_trunc('month', o.created_at) AS month,
       o.seller_id,
       COUNT(DISTINCT o.id) AS orders_count,
       SUM(o.total_cents)  AS total_cents
FROM orders o
WHERE o.payment_status = 'CONFIRMED'
GROUP BY 1,2;`,
		`CREATE OR REPLACE FUNCTION sp_finalize_order(p_order_id bigint)
RETURNS void AS $$
DECLARE
  v_client RECORD;
  v_item RECORD;
  v_discount_rate numeric := 0.0;
  v_credits_added bigint := 0;
BEGIN
  SELECT c.* INTO v_client
  FROM orders o JOIN clients c ON c.id = o.client_id
  WHERE o.id = p_order_id FOR UPDATE;

  IF v_client.supports_flamengo OR v_client.watches_one_piece OR lower(coalesce(v_client.city,'')) = 'sousa' THEN
    v_discount_rate := 0.10;
  END IF;

  UPDATE orders SET subtotal_cents = 0, discount_cents = 0, total_cents = 0
  WHERE id = p_order_id;

  FOR v_item IN
    SELECT oi.*, p.stock, p.amount_credits
    FROM order_items oi JOIN plans p ON p.id = oi.plan_id
    WHERE oi.order_id = p_order_id
  LOOP
    IF v_item.stock < v_item.quantity THEN
      RAISE EXCEPTION 'insufficient stock for plan %', v_item.plan_id;
    END IF;

    UPDATE plans SET stock = stock - v_item.quantity WHERE id = v_item.plan_id;

    UPDATE orders
    SET subtotal_cents = subtotal_cents + (v_item.unit_price_cents * v_item.quantity)
    WHERE id = p_order_id;

    v_credits_added := v_credits_added + (v_item.quantity * v_item.amount_credits);
  END LOOP;

  UPDATE orders
  SET discount_cents = floor(subtotal_cents * v_discount_rate),
      total_cents    = subtotal_cents - discount_cents,
      payment_status = 'CONFIRMED'
  WHERE id = p_order_id;

  INSERT INTO credit_ledger (client_id, type, credits_delta, price_cents_delta, meta)
  SELECT o.client_id, 'TOPUP', v_credits_added, o.total_cents, jsonb_build_object('order_id', o.id)
  FROM orders o WHERE o.id = p_order_id;

  INSERT INTO wallets (client_id, balance_credits) VALUES
    ((SELECT client_id FROM orders WHERE id = p_order_id), v_credits_added)
  ON CONFLICT (client_id)
  DO UPDATE SET balance_credits = wallets.balance_credits + EXCLUDED.balance_credits;
END; $$ LANGUAGE plpgsql;`,
	}

	for i, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("phase 2 statement %d failed: %w", i+1, err)
		}
	}

	var sellerCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sellers`).Scan(&sellerCount); err != nil {
		return fmt.Errorf("failed to count sellers: %w", err)
	}

	if sellerCount == 0 {
		if _, err := db.Exec(`INSERT INTO sellers (name) VALUES ($1), ($2)`, "WebStore", "AdminPanel"); err != nil {
			return fmt.Errorf("failed to seed sellers: %w", err)
		}
	}

	return nil
}
