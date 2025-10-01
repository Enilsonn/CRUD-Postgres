package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
	"github.com/Enilsonn/CRUD-Postgres/internal/service"
	"github.com/Enilsonn/CRUD-Postgres/internal/utils"
	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	type item struct {
		PlanID   int64 `json:"plan_id"`
		Quantity int   `json:"quantity"`
	}
	type req struct {
		ClientID      int64  `json:"client_id"`
		SellerID      int64  `json:"seller_id"`
		PaymentMethod string `json:"payment_method"`
		Items         []item `json:"items"`
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

	items := make([]service.OrderItemRequest, len(payload.Items))
	for idx, it := range payload.Items {
		items[idx] = service.OrderItemRequest{PlanID: it.PlanID, Quantity: it.Quantity}
	}

	order, err := h.service.CreateOrder(r.Context(), service.CreateOrderRequest{
		ClientID:      payload.ClientID,
		SellerID:      payload.SellerID,
		PaymentMethod: payload.PaymentMethod,
		Items:         items,
	})
	if err != nil {
		status := http.StatusInternalServerError
		code := "CREATE_ORDER_FAILED"

		switch {
		case errors.Is(err, repository.ErrOrderWithoutItems):
			status = http.StatusBadRequest
			code = "INVALID_ORDER"
		case strings.Contains(strings.ToLower(err.Error()), "invalid"):
			status = http.StatusBadRequest
			code = "INVALID_ORDER"
		case strings.Contains(strings.ToLower(err.Error()), "not found"):
			status = http.StatusNotFound
			code = "NOT_FOUND"
		case strings.Contains(strings.ToLower(err.Error()), "insufficient stock"):
			status = http.StatusConflict
			code = "INSUFFICIENT_STOCK"
		}

		utils.EncodeJson(w, r, status, map[string]any{
			"error":   true,
			"code":    code,
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusCreated, order)
}

func (h *OrderHandler) FinalizeOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_ORDER_ID",
			"message": "invalid order id",
		})
		return
	}

	resp, err := h.service.FinalizeOrder(r.Context(), orderID)
	if err != nil {
		status := http.StatusInternalServerError
		code := "FINALIZE_ORDER_FAILED"

		switch {
		case errors.Is(err, repository.ErrInsufficientStock):
			status = http.StatusConflict
			code = "INSUFFICIENT_STOCK"
		case errors.Is(err, repository.ErrOrderNotFound):
			status = http.StatusNotFound
			code = "ORDER_NOT_FOUND"
		default:
			status = http.StatusInternalServerError
		}

		utils.EncodeJson(w, r, status, map[string]any{
			"error":   true,
			"code":    code,
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"order":          resp.Order,
		"wallet_balance": resp.WalletBalance,
	})
}

func (h *OrderHandler) ListClientOrders(w http.ResponseWriter, r *http.Request) {
	clientID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_CLIENT_ID",
			"message": "invalid client id",
		})
		return
	}

	orders, err := h.service.ListOrdersByClient(r.Context(), clientID)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "LIST_ORDERS_FAILED",
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, orders)
}
