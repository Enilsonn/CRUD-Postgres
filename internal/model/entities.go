package model

import "time"

type Client struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	Status           bool      `json:"status"`
	RegistrationData time.Time `json:"registration_data"`
}

func NewCliente(name, email, phone string) *Client {
	return &Client{
		Name:             name,
		Email:            email,
		Phone:            phone,
		Status:           true,
		RegistrationData: time.Now(),
	}
}

type Plan struct {
	ID            int64  `json:"id"`
	PlanName      string `json:"plan_name"`
	PriceCents    int64  `json:"price_cents"`
	AmountCredits int    `json:"amount_credits"`
	Status        bool   `json:"status"`
}

func NewPlan(planName string, priceCents int64, amountCredits int) *Plan {
	return &Plan{
		PlanName:      planName,
		PriceCents:    priceCents,
		AmountCredits: amountCredits,
		Status:        true,
	}
}

type Wallet struct {
	ClientID       int64 `json:"client_id"`
	BalanceCredits int64 `json:"balance_credits"`
}

type CreditLedgerEntry struct {
	ID               int64     `json:"id"`
	ClientID         int64     `json:"client_id"`
	Type             string    `json:"type"`
	CreditsDelta     int64     `json:"credits_delta"`
	PriceCentsDelta  int64     `json:"price_cents_delta"`
	Meta             string    `json:"meta"`
	CreatedAt        time.Time `json:"created_at"`
}

type UsageEvent struct {
	ID               int64     `json:"id"`
	ClientID         int64     `json:"client_id"`
	Model            string    `json:"model"`
	PromptTokens     int64     `json:"prompt_tokens"`
	CompletionTokens int64     `json:"completion_tokens"`
	CreditsSpent     int64     `json:"credits_spent"`
	CreatedAt        time.Time `json:"created_at"`
}
