package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Enilsonn/CRUD-Postgres/cmd/configs"
	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
	"github.com/Enilsonn/CRUD-Postgres/internal/service"
	"github.com/spf13/viper"
)

type chatTestCase struct {
	name                string
	model               string
	promptTokens        int64
	completionTokens    int64
	expectedCredits     int64
	pricingRows         *sqlmock.Rows
	walletBalanceBefore int64
	walletBalanceAfter  int64
}

func setupConfig(t *testing.T, host string, defaultModel string) {
	t.Helper()
	viper.Reset()
	viper.Set("ai.ollama_host", host)
	viper.Set("ai.default_model", defaultModel)
	if err := configs.Load("./cmd/main"); err != nil {
		t.Fatalf("load config: %v", err)
	}
}

func TestChatOllamaCredits(t *testing.T) {
	cases := []chatTestCase{
		{
			name:             "matches configured rate",
			model:            "gemma3:1b",
			promptTokens:     900,
			completionTokens: 900,
			expectedCredits:  3,
			pricingRows: sqlmock.NewRows([]string{"pattern", "credits_per_1k_prompt", "credits_per_1k_completion"}).
				AddRow("^gemma3:1b$", 1.0, 2.0),
			walletBalanceBefore: 10,
			walletBalanceAfter:  7,
		},
		{
			name:             "falls back to default pricing",
			model:            "unknown-model",
			promptTokens:     900,
			completionTokens: 900,
			expectedCredits:  2,
			pricingRows: sqlmock.NewRows([]string{"pattern", "credits_per_1k_prompt", "credits_per_1k_completion"}).
				AddRow("^doesnotmatch$", 3.0, 3.0),
			walletBalanceBefore: 5,
			walletBalanceAfter:  3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock new: %v", err)
			}
			defer db.Close()

			mock.ExpectQuery(`(?s)SELECT pattern, credits_per_1k_prompt, credits_per_1k_completion FROM model_pricing`).
				WillReturnRows(tc.pricingRows)

			mock.ExpectBegin()
			mock.ExpectQuery(`(?s)SELECT COALESCE\(balance_credits, 0\) FROM wallets`).
				WithArgs(int64(1)).
				WillReturnRows(sqlmock.NewRows([]string{"balance_credits"}).AddRow(tc.walletBalanceBefore))
			mock.ExpectExec(`(?s)INSERT INTO usage_events`).
				WithArgs(int64(1), tc.model, tc.promptTokens, tc.completionTokens, tc.expectedCredits).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec(`(?s)INSERT INTO credit_ledger`).
				WithArgs(int64(1), -tc.expectedCredits, sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(`(?s)UPDATE wallets`).
				WithArgs(int64(1), tc.expectedCredits).
				WillReturnRows(sqlmock.NewRows([]string{"balance_credits"}).AddRow(tc.walletBalanceAfter))
			mock.ExpectCommit()

			walletRepo := repository.NewWalletRepository(db)
			pricingRepo := repository.NewPricingRepository(db)
			pricingSvc := service.NewPricingService(pricingRepo)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/chat" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}

				resp := map[string]any{
					"model": tc.model,
					"message": map[string]any{
						"role":    "assistant",
						"content": "hi",
					},
					"done":                 true,
					"prompt_eval_count":    tc.promptTokens,
					"eval_count":           tc.completionTokens,
					"total_duration":       0,
					"eval_duration":        0,
					"prompt_eval_duration": 0,
				}
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			setupConfig(t, server.URL, tc.model)

			handler := NewChatHandler(walletRepo, pricingSvc)
			handler.HTTPClient = server.Client()

			payload := map[string]any{
				"client_id": int64(1),
				"model":     tc.model,
				"messages": []map[string]string{
					{"role": "user", "content": "hello"},
				},
			}
			body := new(bytes.Buffer)
			if err := json.NewEncoder(body).Encode(payload); err != nil {
				t.Fatalf("encode payload: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/chat/ollama", body)
			rr := httptest.NewRecorder()

			handler.ChatOllama(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("unexpected status: %d body: %s", rr.Code, rr.Body.String())
			}

			var resp ChatResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if resp.CreditsCharged != tc.expectedCredits {
				t.Fatalf("expected credits %d got %d", tc.expectedCredits, resp.CreditsCharged)
			}

			if resp.WalletBalance != tc.walletBalanceAfter {
				t.Fatalf("expected wallet balance %d got %d", tc.walletBalanceAfter, resp.WalletBalance)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}
