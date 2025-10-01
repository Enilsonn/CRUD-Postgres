package controller

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
	"github.com/Enilsonn/CRUD-Postgres/internal/service"
	"github.com/Enilsonn/CRUD-Postgres/internal/utils"
	"github.com/go-chi/chi/v5"
)

type WalletHandler struct {
	WalletRepo *repository.WalletRepository
	PlanRepo   *repository.ProductRepository
	PricingSvc *service.PricingService
}

func NewWalletHandler(walletRepo *repository.WalletRepository, planRepo *repository.ProductRepository, pricingSvc *service.PricingService) *WalletHandler {
	return &WalletHandler{
		WalletRepo: walletRepo,
		PlanRepo:   planRepo,
		PricingSvc: pricingSvc,
	}
}

func (h *WalletHandler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	clientID, err := strconv.ParseInt(chi.URLParam(r, "client_id"), 10, 64)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"code":    "INVALID_CLIENT_ID",
				"message": "invalid client ID",
			})
		return
	}

	wallet, err := h.WalletRepo.GetWalletByClientID(ctx, clientID)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound,
			map[string]any{
				"error":   true,
				"code":    "WALLET_NOT_FOUND",
				"message": "wallet not found",
			})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, wallet)
}

func (h *WalletHandler) GetLedgerEntries(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	clientID, err := strconv.ParseInt(chi.URLParam(r, "client_id"), 10, 64)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"code":    "INVALID_CLIENT_ID",
				"message": "invalid client ID",
			})
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	entries, err := h.WalletRepo.GetLedgerEntries(ctx, clientID, limit, offset)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError,
			map[string]any{
				"error":   true,
				"code":    "DATABASE_ERROR",
				"message": err.Error(),
			})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, entries)
}

func (h *WalletHandler) TopUpCredits(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	clientID, err := strconv.ParseInt(chi.URLParam(r, "client_id"), 10, 64)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"code":    "INVALID_CLIENT_ID",
				"message": "invalid client ID",
			})
		return
	}

	type req struct {
		PlanID    int64  `json:"plan_id"`
		RequestID string `json:"request_id,omitempty"`
	}

	topup, err := utils.DecodeJson[req](r)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("invalid request body: %v", err),
			})
		return
	}

	// Get plan details
	plan, err := h.PlanRepo.GetProductByID(ctx, topup.PlanID)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound,
			map[string]any{
				"error":   true,
				"code":    "PLAN_NOT_FOUND",
				"message": "plan not found",
			})
		return
	}

	// Process top-up transaction
	newBalance, err := h.WalletRepo.ProcessTopUp(ctx, clientID, plan.AmountCredits, plan.PriceCents, topup.RequestID)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError,
			map[string]any{
				"error":   true,
				"code":    "TOPUP_FAILED",
				"message": err.Error(),
			})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK,
		map[string]any{
			"client_id":       clientID,
			"balance_credits": newBalance,
			"added_credits":   plan.AmountCredits,
			"cost_cents":      plan.PriceCents,
		})
}

func (h *WalletHandler) ProcessUsage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	type req struct {
		ClientID         int64  `json:"client_id"`
		Model            string `json:"model"`
		PromptTokens     int64  `json:"prompt_tokens"`
		CompletionTokens int64  `json:"completion_tokens"`
	}

	usage, err := utils.DecodeJson[req](r)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("invalid request body: %v", err),
			})
		return
	}

	// Calculate credits to deduct, using dynamic pricing when available
	totalTokens := usage.PromptTokens + usage.CompletionTokens
	creditsNeeded := int64(math.Ceil(float64(totalTokens) / 1000.0))
	meta := map[string]any{
		"model": usage.Model,
		"ppk":   1.0,
		"cpk":   1.0,
	}

	if h.PricingSvc != nil {
		calcCredits, ppk, cpk, err := h.PricingSvc.ComputeCredits(ctx, usage.Model, usage.PromptTokens, usage.CompletionTokens)
		if err != nil {
			utils.EncodeJson(w, r, http.StatusInternalServerError,
				map[string]any{
					"error":   true,
					"code":    "PRICING_FAILED",
					"message": err.Error(),
				})
			return
		}
		creditsNeeded = calcCredits
		meta["ppk"] = ppk
		meta["cpk"] = cpk
	}

	// Process usage transaction
	newBalance, err := h.WalletRepo.ProcessUsage(ctx, usage.ClientID, usage.Model, usage.PromptTokens, usage.CompletionTokens, creditsNeeded, meta)
	if err != nil {
		if err.Error() == "insufficient credits" {
			utils.EncodeJson(w, r, http.StatusConflict,
				map[string]any{
					"error":   true,
					"code":    "INSUFFICIENT_CREDITS",
					"message": "not enough credits",
				})
			return
		}
		utils.EncodeJson(w, r, http.StatusInternalServerError,
			map[string]any{
				"error":   true,
				"code":    "USAGE_FAILED",
				"message": err.Error(),
			})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK,
		map[string]any{
			"client_id":        usage.ClientID,
			"balance_credits":  newBalance,
			"credits_spent":    creditsNeeded,
			"tokens_processed": totalTokens,
		})
}
