package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
	"github.com/Enilsonn/CRUD-Postgres/internal/utils"
	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	Repo *repository.ProductRepository
}

func NewProductHandler(repo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{
		Repo: repo,
	}
}

func (h *ProductHandler) CreateClientProduct(w http.ResponseWriter, r *http.Request) {
	type req struct {
		PlanName      string `json:"plan_name"`
		PriceCents    int64  `json:"price_cents"`
		AmountCredits int    `json:"amount_credits"`
	}
	plan, err := utils.DecodeJson[req](r)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("invalid request body: %v", err),
			})
		return
	}

	planCompleted := model.NewPlan(plan.PlanName, plan.PriceCents, plan.AmountCredits)

	id, err := h.Repo.CreateClientProduct(*planCompleted)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError,
			map[string]any{
				"error":   true,
				"code":    "DATABASE_ERROR",
				"message": err.Error(),
			})
		return
	}

	planCompleted.ID = id

	utils.EncodeJson(w, r, http.StatusCreated, planCompleted)
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	product, err := h.Repo.GetProductByID(int64(id))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK,
		product,
	)
}

func (h *ProductHandler) GetClientProductByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"message": "name must be passed",
			})
		return
	}

	product, err := h.Repo.GetClientProductByName(name)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK,
		product,
	)
}

func (h *ProductHandler) GetAllClientProduct(w http.ResponseWriter, r *http.Request) {
	products, err := h.Repo.GetAllClientProduct()
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK,
		products,
	)
}

func (h *ProductHandler) UpdateClientProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	product, err := utils.DecodeJson[model.Plan](r)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	rowsAffected, err := h.Repo.UpdateClientProduct(int64(id), product)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	if rowsAffected == 1 {
		utils.EncodeJson(w, r, http.StatusOK,
			map[string]any{
				"error":   false,
				"message": fmt.Sprintf("product %d updated successfully", id),
			},
		)
	} else {
		utils.EncodeJson(w, r, http.StatusOK,
			map[string]any{
				"error":   false,
				"message": fmt.Sprintf("%d products updated successfully", rowsAffected),
			},
		)
	}
}

func (h *ProductHandler) DeleteClientProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	if err = h.Repo.DeleteClientProduct(int64(id)); err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	utils.EncodeJson(w, r, http.StatusNoContent, map[string]any{})
}
