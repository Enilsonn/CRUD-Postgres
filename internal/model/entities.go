package model

import (
	"encoding/json"
	"time"
)

type Client struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	Status           bool      `json:"status"`
	RegistrationData time.Time `json:"registration_data"`
	SupportsFlamengo bool      `json:"supports_flamengo"`
	WatchesOnePiece  bool      `json:"watches_one_piece"`
	City             string    `json:"city"`
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
	ID                 int64  `json:"id"`
	PlanName           string `json:"plan_name"`
	PriceCents         int64  `json:"price_cents"`
	AmountCredits      int    `json:"amount_credits"`
	Status             bool   `json:"status"`
	Category           string `json:"category"`
	ManufacturedInMari bool   `json:"manufactured_in_mari"`
	Stock              int    `json:"stock"`
}

func NewPlan(planName string, priceCents int64, amountCredits int) *Plan {
	return &Plan{
		PlanName:      planName,
		PriceCents:    priceCents,
		AmountCredits: amountCredits,
		Status:        true,
		Category:      "CREDITS",
		Stock:         999999,
	}
}

type Wallet struct {
	ClientID       int64 `json:"client_id"`
	BalanceCredits int64 `json:"balance_credits"`
}

type CreditLedgerEntry struct {
	ID              int64           `json:"id"`
	ClientID        int64           `json:"client_id"`
	Type            string          `json:"type"`
	CreditsDelta    int64           `json:"credits_delta"`
	PriceCentsDelta int64           `json:"price_cents_delta"`
	Meta            json.RawMessage `json:"meta"`
	CreatedAt       time.Time       `json:"created_at"`
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

type Seller struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Order struct {
	ID            int64       `json:"id"`
	ClientID      int64       `json:"client_id"`
	SellerID      int64       `json:"seller_id"`
	CreatedAt     time.Time   `json:"created_at"`
	PaymentMethod string      `json:"payment_method"`
	PaymentStatus string      `json:"payment_status"`
	SubtotalCents int64       `json:"subtotal_cents"`
	DiscountCents int64       `json:"discount_cents"`
	TotalCents    int64       `json:"total_cents"`
	Items         []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID             int64 `json:"id"`
	OrderID        int64 `json:"order_id"`
	PlanID         int64 `json:"plan_id"`
	Quantity       int   `json:"quantity"`
	UnitPriceCents int64 `json:"unit_price_cents"`
}

type SellerMonthlySales struct {
	Month       time.Time `json:"month"`
	SellerID    int64     `json:"seller_id"`
	OrdersCount int64     `json:"orders_count"`
	TotalCents  int64     `json:"total_cents"`
}
