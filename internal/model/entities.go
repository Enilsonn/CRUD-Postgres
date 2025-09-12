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

type ClientProduct struct {
	ID            int     `json:"id"`
	PlanName      string  `json:"plan_name"`
	PriceCents    float32 `json:"price_cents"`
	AmountCredits int     `json:"amount_credits"`
	Status        bool    `json:"status"`
}

func NewClientProduct(plan_name string, price_cents float32, amount_credites int) *ClientProduct {
	return &ClientProduct{
		PlanName:      plan_name,
		PriceCents:    price_cents,
		AmountCredits: amount_credites,
		Status:        true,
	}
}
