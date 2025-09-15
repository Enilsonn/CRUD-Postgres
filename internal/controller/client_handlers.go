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
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	client, err := utils.DecodeJson[req](r)
	if err != nil {
		utils.EncodeJson(w, r,
			http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"message": fmt.Sprintf("invalid request body: %v", err),
			})
		return
	}
	// não haverá verificação se os campos são válidos

	clientCompleted := model.NewCliente(client.Name, client.Email, client.Phone)

	id, err := h.Repo.CreateClient(*clientCompleted)
	if err != nil {
		utils.EncodeJson(w, r,
			http.StatusInternalServerError,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	clientCompleted.ID = id

	utils.EncodeJson(w, r, http.StatusCreated, clientCompleted)
}

func (h *ClientHandler) GetClientByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"message": fmt.Sprintf("invalid id: %v", err),
			})
		return
	}

	client, err := h.Repo.GetClientByID(int64(id))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK,
		client,
	)
}

func (h *ClientHandler) GetClientByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"message": "name must be passed",
			})
		return
	}

	clients, err := h.Repo.GetClientByName(name)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError,
			map[string]any{
				"error":   true,
				"code":    "DATABASE_ERROR",
				"message": err.Error(),
			})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, clients)
}

func (h *ClientHandler) GetAllClients(w http.ResponseWriter, r *http.Request) {
	clients, err := h.Repo.GetAllClients()
	if err != nil {
		utils.EncodeJson(w, r, http.StatusNotFound,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK,
		clients)
}

func (h *ClientHandler) UpdateClients(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	// não impede que o cliente passe id ou status (o que não é permitido a ele)
	client, err := utils.DecodeJson[model.Client](r)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	rowsAffected, err := h.Repo.UpdateClients(int64(id), client)
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
				"message": fmt.Sprintf("client %d updated successfully", id),
			},
		)
	} else {
		utils.EncodeJson(w, r, http.StatusOK,
			map[string]any{
				"error":   false,
				"message": fmt.Sprintf("%d clients updated successfully", rowsAffected),
			},
		)
	}
}

func (h *ClientHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	if err = h.Repo.DeleteClient(int64(id)); err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError,
			map[string]any{
				"error":   true,
				"message": err,
			})
		return
	}

	utils.EncodeJson(w, r, http.StatusNoContent, map[string]any{})

}
