package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
	"github.com/Enilsonn/CRUD-Postgres/internal/utils"
	"github.com/go-chi/chi/v5"
)

type ClientHandler struct {
	Repo *repository.ClientRepository
}

func NewClientHandler(repo *repository.ClientRepository) *ClientHandler {
	return &ClientHandler{
		Repo: repo,
	}
}

func (h *ClientHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Name             string  `json:"name"`
		Email            string  `json:"email"`
		Phone            string  `json:"phone"`
		SupportsFlamengo *bool   `json:"supports_flamengo"`
		WatchesOnePiece  *bool   `json:"watches_one_piece"`
		City             *string `json:"city"`
	}

	payload, err := utils.DecodeJson[req](r)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"message": fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}

	client := model.NewCliente(payload.Name, payload.Email, payload.Phone)
	if payload.SupportsFlamengo != nil {
		client.SupportsFlamengo = *payload.SupportsFlamengo
	}
	if payload.WatchesOnePiece != nil {
		client.WatchesOnePiece = *payload.WatchesOnePiece
	}
	if payload.City != nil {
		trimmed := strings.TrimSpace(*payload.City)
		client.City = sql.NullString{String: trimmed, Valid: trimmed != ""}
	}

	id, err := h.Repo.CreateClient(r.Context(), *client)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	client.ID = id

	utils.EncodeJson(w, r, http.StatusCreated, client)
}

func (h *ClientHandler) GetClientByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"message": fmt.Sprintf("invalid id: %v", err),
		})
		return
	}

	client, err := h.Repo.GetClientByID(r.Context(), int64(id))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, client)
}

func (h *ClientHandler) GetClientByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"message": "name must be passed",
		})
		return
	}

	clients, err := h.Repo.GetClientByName(r.Context(), name)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "DATABASE_ERROR",
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, clients)
}

func (h *ClientHandler) GetAllClients(w http.ResponseWriter, r *http.Request) {
	clients, err := h.Repo.GetAllClients(r.Context())
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, clients)
}

func (h *ClientHandler) UpdateClients(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	current, err := h.Repo.GetClientByID(r.Context(), int64(id))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	type req struct {
		Name             *string `json:"name"`
		Email            *string `json:"email"`
		Phone            *string `json:"phone"`
		SupportsFlamengo *bool   `json:"supports_flamengo"`
		WatchesOnePiece  *bool   `json:"watches_one_piece"`
		City             *string `json:"city"`
	}

	payload, err := utils.DecodeJson[req](r)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	if payload.Name != nil {
		current.Name = *payload.Name
	}
	if payload.Email != nil {
		current.Email = *payload.Email
	}
	if payload.Phone != nil {
		current.Phone = *payload.Phone
	}
	if payload.SupportsFlamengo != nil {
		current.SupportsFlamengo = *payload.SupportsFlamengo
	}
	if payload.WatchesOnePiece != nil {
		current.WatchesOnePiece = *payload.WatchesOnePiece
	}
	if payload.City != nil {
		trimmed := strings.TrimSpace(*payload.City)
		current.City = sql.NullString{String: trimmed, Valid: trimmed != ""}
	}

	rowsAffected, err := h.Repo.UpdateClients(r.Context(), int64(id), *current)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	if rowsAffected == 1 {
		utils.EncodeJson(w, r, http.StatusOK, current)
	} else {
		utils.EncodeJson(w, r, http.StatusOK, map[string]any{
			"error":   false,
			"message": fmt.Sprintf("%d clients updated successfully", rowsAffected),
		})
	}
}

func (h *ClientHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	if err = h.Repo.DeleteClient(r.Context(), int64(id)); err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusNoContent, map[string]any{})
}
