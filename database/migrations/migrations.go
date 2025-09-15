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

	return nil
}
