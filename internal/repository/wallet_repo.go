package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) GetWalletByClientID(ctx context.Context, clientID int64) (*model.Wallet, error) {
	wallet := &model.Wallet{}

	sqlGet := `
		SELECT client_id, balance_credits
		FROM wallets
		WHERE client_id = $1`
	err := r.db.QueryRowContext(ctx, sqlGet,
		clientID).Scan(&wallet.ClientID, &wallet.BalanceCredits)

	sqlInsert := `
			INSERT INTO wallets (client_id, balance_credits)
			VALUES ($1, 0)
			ON CONFLICT (client_id) DO NOTHING`
	if err == sql.ErrNoRows {
		// Create wallet if it doesn't exist
		_, err = r.db.ExecContext(ctx, sqlInsert,
			clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to create wallet: %w", err)
		}

		wallet.ClientID = clientID
		wallet.BalanceCredits = 0
		return wallet, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet, nil
}

func (r *WalletRepository) GetLedgerEntries(ctx context.Context, clientID int64, limit, offset int) ([]*model.CreditLedgerEntry, error) {
	sql := `
		SELECT id, client_id, type, credits_delta, price_cents_delta, meta, created_at
		FROM credit_ledger
		WHERE client_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx,
		sql,
		clientID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to query ledger entries: %w", err)
	}
	defer rows.Close()

	var entries []*model.CreditLedgerEntry
	for rows.Next() {
		entry := &model.CreditLedgerEntry{}
		var metaBytes []byte
		err := rows.Scan(
			&entry.ID,
			&entry.ClientID,
			&entry.Type,
			&entry.CreditsDelta,
			&entry.PriceCentsDelta,
			&metaBytes,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ledger entry: %w", err)
		}
		entry.Meta = json.RawMessage(metaBytes)
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *WalletRepository) ProcessTopUp(ctx context.Context, clientID int64, credits int, priceCents int64, requestID string) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create metadata
	meta := map[string]interface{}{
		"request_id": requestID,
		"type":       "plan_purchase",
	}
	metaBytes, _ := json.Marshal(meta)

	// Insert into ledger
	sql := `
		INSERT INTO credit_ledger (client_id, type, credits_delta, price_cents_delta, meta)
		VALUES ($1, 'TOPUP', $2, $3, $4)`
	_, err = tx.ExecContext(ctx,
		sql,
		clientID, int64(credits), priceCents, metaBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to insert ledger entry: %w", err)
	}

	// Upsert wallet balance
	var newBalance int64
	sql = `
		INSERT INTO wallets (client_id, balance_credits)
		VALUES ($1, $2)
		ON CONFLICT (client_id)
		DO UPDATE SET balance_credits = wallets.balance_credits + $2
		RETURNING balance_credits`

	err = tx.QueryRowContext(ctx,
		sql,
		clientID, int64(credits)).Scan(&newBalance)
	if err != nil {
		return 0, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newBalance, nil
}

func (r *WalletRepository) ProcessUsage(ctx context.Context, clientID int64, model string, promptTokens, completionTokens, creditsSpent int64) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check current balance
	var currentBalance int64
	sqlQuerie := `
		SELECT COALESCE(balance_credits, 0)
		FROM wallets
		WHERE client_id = $1`
	err = tx.QueryRowContext(ctx,
		sqlQuerie,
		clientID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("wallet not found")
	} else if err != nil {
		return 0, fmt.Errorf("failed to check wallet balance: %w", err)
	}

	// Check if sufficient balance
	if currentBalance < creditsSpent {
		return 0, fmt.Errorf("insufficient credits")
	}

	// Insert usage event
	sqlQuerie = `
		INSERT INTO usage_events (client_id, model, prompt_tokens, completion_tokens, credits_spent)
		VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.ExecContext(ctx,
		sqlQuerie,
		clientID, model, promptTokens, completionTokens, creditsSpent)
	if err != nil {
		return 0, fmt.Errorf("failed to insert usage event: %w", err)
	}

	// Create metadata
	meta := map[string]interface{}{
		"model":             model,
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
	}
	metaBytes, _ := json.Marshal(meta)

	// Insert into ledger (negative for usage)
	sqlQuerie = `
		INSERT INTO credit_ledger (client_id, type, credits_delta, price_cents_delta, meta)
		VALUES ($1, 'USAGE', $2, 0, $3)`
	_, err = tx.ExecContext(ctx,
		sqlQuerie,
		clientID, -creditsSpent, metaBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to insert ledger entry: %w", err)
	}

	// Update wallet balance
	var newBalance int64
	sqlQuerie = `
		UPDATE wallets
		SET balance_credits = balance_credits - $2
		WHERE client_id = $1
		RETURNING balance_credits`
	err = tx.QueryRowContext(ctx,
		sqlQuerie,
		clientID, creditsSpent).Scan(&newBalance)
	if err != nil {
		return 0, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newBalance, nil
}
