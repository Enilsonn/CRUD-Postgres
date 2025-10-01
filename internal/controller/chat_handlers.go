package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/Enilsonn/CRUD-Postgres/cmd/configs"
	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
	"github.com/Enilsonn/CRUD-Postgres/internal/service"
	"github.com/Enilsonn/CRUD-Postgres/internal/utils"
)

// ChatHandler provides endpoints backed by local Ollama chat models.
type ChatHandler struct {
	WalletRepo *repository.WalletRepository
	HTTPClient *http.Client
	PricingSvc *service.PricingService
}

// NewChatHandler wires the dependencies required for chat operations.
func NewChatHandler(walletRepo *repository.WalletRepository, pricingSvc *service.PricingService) *ChatHandler {
	return &ChatHandler{
		WalletRepo: walletRepo,
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
		PricingSvc: pricingSvc,
	}
}

// ChatMessage represents a single role/content pair in the chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the incoming payload for the Ollama chat endpoint.
type ChatRequest struct {
	ClientID int64          `json:"client_id"`
	Model    string         `json:"model,omitempty"`
	Messages []ChatMessage  `json:"messages"`
	Options  map[string]any `json:"options,omitempty"`
	Stream   *bool          `json:"stream,omitempty"`
}

// ChatResponse is the successful response returned to API callers.
type ChatResponse struct {
	Model            string `json:"model"`
	Reply            string `json:"reply"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
	CreditsCharged   int64  `json:"credits_charged"`
	WalletBalance    int64  `json:"wallet_balance"`
}

// ollamaMessage mirrors Ollama's message schema.
type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaChatRequest is the payload sent to the Ollama HTTP API.
type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Options  map[string]any  `json:"options,omitempty"`
	Stream   bool            `json:"stream"`
}

// ollamaChatResponse represents the non-streaming response from Ollama.
type ollamaChatResponse struct {
	Model              string        `json:"model"`
	Message            ollamaMessage `json:"message"`
	Done               bool          `json:"done"`
	PromptEvalCount    int64         `json:"prompt_eval_count"`
	EvalCount          int64         `json:"eval_count"`
	TotalDuration      int64         `json:"total_duration,omitempty"`
	EvalDuration       int64         `json:"eval_duration,omitempty"`
	PromptEvalDuration int64         `json:"prompt_eval_duration,omitempty"`
}

// ChatOllama handles POST /api/chat/ollama requests.
func (h *ChatHandler) ChatOllama(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_REQUEST",
			"message": fmt.Sprintf("invalid body: %v", err),
		})
		return
	}

	if req.ClientID == 0 || len(req.Messages) == 0 {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "INVALID_REQUEST",
			"message": "client_id and messages are required",
		})
		return
	}

	ai := configs.GetAI()

	model := req.Model
	if model == "" {
		model = ai.DefaultModel
	}

	if req.Stream != nil && *req.Stream {
		utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":   true,
			"code":    "STREAMING_NOT_SUPPORTED",
			"message": "stream=true is not supported",
		})
		return
	}

	oMessages := make([]ollamaMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		oMessages = append(oMessages, ollamaMessage{Role: msg.Role, Content: msg.Content})
	}

	payload := ollamaChatRequest{
		Model:    model,
		Messages: oMessages,
		Options:  req.Options,
		Stream:   false,
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "REQUEST_BUILD_FAILED",
			"message": err.Error(),
		})
		return
	}

	oURL := fmt.Sprintf("%s/api/chat", ai.OllamaHost)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, oURL, buf)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "REQUEST_BUILD_FAILED",
			"message": err.Error(),
		})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.HTTPClient.Do(httpReq)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "OLLAMA_UNAVAILABLE",
			"message": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusBadGateway, map[string]any{
			"error":   true,
			"code":    "OLLAMA_READ_ERROR",
			"message": err.Error(),
		})
		return
	}

	if resp.StatusCode >= http.StatusMultipleChoices {
		utils.EncodeJson(w, r, http.StatusBadGateway, map[string]any{
			"error":  true,
			"code":   "OLLAMA_ERROR",
			"status": resp.StatusCode,
			"body":   string(body),
		})
		return
	}

	var oResp ollamaChatResponse
	if err := json.Unmarshal(body, &oResp); err != nil {
		utils.EncodeJson(w, r, http.StatusBadGateway, map[string]any{
			"error":   true,
			"code":    "OLLAMA_PARSE_ERROR",
			"message": err.Error(),
		})
		return
	}

	promptTokens := oResp.PromptEvalCount
	completionTokens := oResp.EvalCount
	totalTokens := promptTokens + completionTokens
	var (
		credits int64
		ppk     float64 = 1.0
		cpk     float64 = 1.0
	)

	if h.PricingSvc != nil {
		credits, ppk, cpk, err = h.PricingSvc.ComputeCredits(ctx, model, promptTokens, completionTokens)
		if err != nil {
			utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
				"error":   true,
				"code":    "PRICING_FAILED",
				"message": err.Error(),
			})
			return
		}
	} else {
		credits = int64(math.Ceil(float64(totalTokens) / 1000.0))
	}

	meta := map[string]any{
		"model": model,
		"ppk":   ppk,
		"cpk":   cpk,
	}

	newBalance, err := h.WalletRepo.ProcessUsage(ctx, req.ClientID, model, promptTokens, completionTokens, credits, meta)
	if err != nil {
		switch err.Error() {
		case "insufficient credits":
			utils.EncodeJson(w, r, http.StatusConflict, map[string]any{
				"error":   true,
				"code":    "INSUFFICIENT_CREDITS",
				"message": "not enough credits",
			})
			return
		case "wallet not found":
			utils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
				"error":   true,
				"code":    "UNKNOWN_CLIENT",
				"message": "unknown client",
			})
			return
		default:
			utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
				"error":   true,
				"code":    "USAGE_DEBIT_FAILED",
				"message": err.Error(),
			})
			return
		}
	}

	respBody := ChatResponse{
		Model:            model,
		Reply:            oResp.Message.Content,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		CreditsCharged:   credits,
		WalletBalance:    newBalance,
	}

	utils.EncodeJson(w, r, http.StatusOK, respBody)
}
