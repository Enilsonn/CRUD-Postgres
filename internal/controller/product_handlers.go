package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
	"github.com/Enilsonn/CRUD-Postgres/internal/service"
	"github.com/Enilsonn/CRUD-Postgres/internal/utils"
	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	Repo        *repository.ProductRepository
	PlanService *service.PlanService
}

func NewProductHandler(repo *repository.ProductRepository, planService *service.PlanService) *ProductHandler {
	return &ProductHandler{
		Repo:        repo,
		PlanService: planService,
	}
}

func (h *ProductHandler) CreateClientProduct(w http.ResponseWriter, r *http.Request) {
	type req struct {
		PlanName           string  `json:"plan_name"`
		PriceCents         int64   `json:"price_cents"`
		AmountCredits      int     `json:"amount_credits"`
		Category           *string `json:"category"`
		ManufacturedInMari *bool   `json:"manufactured_in_mari"`
		Stock              *int    `json:"stock"`
	}

	payload, err := utils.DecodeJson[req](r)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_REQUEST",
			"message": fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}

	plan := model.NewPlan(payload.PlanName, payload.PriceCents, payload.AmountCredits)
	if payload.Category != nil && strings.TrimSpace(*payload.Category) != "" {
		plan.Category = strings.TrimSpace(*payload.Category)
	}
	if payload.ManufacturedInMari != nil {
		plan.ManufacturedInMari = *payload.ManufacturedInMari
	}
	if payload.Stock != nil {
		if *payload.Stock < 0 {
			utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
				"error":   true,
				"code":    "INVALID_STOCK",
				"message": "stock must be non-negative",
			})
			return
		}
		plan.Stock = *payload.Stock
	}

	id, err := h.Repo.CreateClientProduct(r.Context(), *plan)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "DATABASE_ERROR",
			"message": err.Error(),
		})
		return
	}

	plan.ID = id

	utils.EncodeJson(w, r, http.StatusCreated, plan)
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_PLAN_ID",
			"message": "invalid plan id",
		})
		return
	}

	product, err := h.Repo.GetProductByID(r.Context(), id)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
			"error":   true,
			"code":    "PLAN_NOT_FOUND",
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, product)
}

func (h *ProductHandler) GetClientProductByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_NAME",
			"message": "name must be passed",
		})
		return
	}

	product, err := h.Repo.GetClientProductByName(r.Context(), name)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
			"error":   true,
			"code":    "PLAN_NOT_FOUND",
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, product)
}

func (h *ProductHandler) GetAllClientProduct(w http.ResponseWriter, r *http.Request) {
	products, err := h.Repo.GetAllClientProduct(r.Context())
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "DATABASE_ERROR",
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, products)
}

func (h *ProductHandler) UpdateClientProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_PLAN_ID",
			"message": "invalid plan id",
		})
		return
	}

	current, err := h.Repo.GetProductByID(r.Context(), id)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
			"error":   true,
			"code":    "PLAN_NOT_FOUND",
			"message": err.Error(),
		})
		return
	}

	type req struct {
		PlanName           *string `json:"plan_name"`
		PriceCents         *int64  `json:"price_cents"`
		AmountCredits      *int    `json:"amount_credits"`
		Category           *string `json:"category"`
		ManufacturedInMari *bool   `json:"manufactured_in_mari"`
		Stock              *int    `json:"stock"`
	}

	payload, err := utils.DecodeJson[req](r)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_REQUEST",
			"message": fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}

	if payload.PlanName != nil {
		current.PlanName = *payload.PlanName
	}
	if payload.PriceCents != nil {
		if *payload.PriceCents < 0 {
			utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
				"error":   true,
				"code":    "INVALID_PRICE",
				"message": "price_cents must be non-negative",
			})
			return
		}
		current.PriceCents = *payload.PriceCents
	}
	if payload.AmountCredits != nil {
		if *payload.AmountCredits <= 0 {
			utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
				"error":   true,
				"code":    "INVALID_AMOUNT",
				"message": "amount_credits must be positive",
			})
			return
		}
		current.AmountCredits = *payload.AmountCredits
	}
	if payload.Category != nil {
		trimmed := strings.TrimSpace(*payload.Category)
		if trimmed != "" {
			current.Category = trimmed
		}
	}
	if payload.ManufacturedInMari != nil {
		current.ManufacturedInMari = *payload.ManufacturedInMari
	}
	if payload.Stock != nil {
		if *payload.Stock < 0 {
			utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
				"error":   true,
				"code":    "INVALID_STOCK",
				"message": "stock must be non-negative",
			})
			return
		}
		current.Stock = *payload.Stock
	}

	if _, err := h.Repo.UpdateClientProduct(r.Context(), id, *current); err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "DATABASE_ERROR",
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, current)
}

func (h *ProductHandler) DeleteClientProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_PLAN_ID",
			"message": "invalid plan id",
		})
		return
	}

	if err := h.Repo.DeleteClientProduct(r.Context(), id); err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "DATABASE_ERROR",
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusNoContent, map[string]any{})
}

func (h *ProductHandler) SearchPlans(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filters := service.PlanSearchFilters{}

	filters.Name = strings.TrimSpace(query.Get("name"))

	if v := strings.TrimSpace(query.Get("min_price_cents")); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
				"error":   true,
				"code":    "INVALID_MIN_PRICE",
				"message": "min_price_cents must be an integer",
			})
			return
		}
		filters.MinPriceCents = &parsed
	}

	if v := strings.TrimSpace(query.Get("max_price_cents")); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
				"error":   true,
				"code":    "INVALID_MAX_PRICE",
				"message": "max_price_cents must be an integer",
			})
			return
		}
		filters.MaxPriceCents = &parsed
	}

	filters.Category = strings.TrimSpace(query.Get("category"))

	if v := strings.TrimSpace(query.Get("manufactured_in_mari")); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
				"error":   true,
				"code":    "INVALID_MANUFACTURED_IN_MARI",
				"message": "manufactured_in_mari must be a boolean",
			})
			return
		}
		filters.ManufacturedInMari = &parsed
	}

	plans, err := h.PlanService.Search(r.Context(), filters)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "SEARCH_FAILED",
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, plans)
}

func (h *ProductHandler) LowStock(w http.ResponseWriter, r *http.Request) {
	plans, err := h.PlanService.LowStock(r.Context())
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "LOW_STOCK_FAILED",
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, plans)
}
