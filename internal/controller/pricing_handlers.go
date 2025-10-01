package controller

import (
	"context"
	"math"
	"net/http"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
	"github.com/Enilsonn/CRUD-Postgres/internal/utils"
)

type PricingHandler struct {
	Repo *repository.PricingRepository
}

func NewPricingHandler(repo *repository.PricingRepository) *PricingHandler {
	return &PricingHandler{Repo: repo}
}

func (h *PricingHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	rates, err := h.Repo.ListActive(ctx)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "PRICING_LIST_FAILED",
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, rates)
}

func (h *PricingHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	type req struct {
		ID                     int64   `json:"id,omitempty"`
		Pattern                string  `json:"pattern"`
		CreditsPer1KPrompt     float64 `json:"credits_per_1k_prompt"`
		CreditsPer1KCompletion float64 `json:"credits_per_1k_completion"`
		Priority               int     `json:"priority,omitempty"`
		Active                 *bool   `json:"active,omitempty"`
	}

	payload, err := utils.DecodeJson[req](r)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	if payload.Pattern == "" {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_PATTERN",
			"message": "pattern is required",
		})
		return
	}

	if payload.CreditsPer1KPrompt <= 0 || payload.CreditsPer1KCompletion <= 0 || math.IsNaN(payload.CreditsPer1KPrompt) || math.IsNaN(payload.CreditsPer1KCompletion) {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_RATE",
			"message": "rates must be positive",
		})
		return
	}

	priority := payload.Priority
	if priority == 0 {
		priority = 100
	}

	active := true
	if payload.Active != nil {
		active = *payload.Active
	}

	rate := &model.ModelPricing{
		ID:                     payload.ID,
		Pattern:                payload.Pattern,
		CreditsPer1KPrompt:     payload.CreditsPer1KPrompt,
		CreditsPer1KCompletion: payload.CreditsPer1KCompletion,
		Priority:               priority,
		Active:                 active,
	}

	saved, err := h.Repo.Upsert(ctx, rate)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "PRICING_SAVE_FAILED",
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, saved)
}
