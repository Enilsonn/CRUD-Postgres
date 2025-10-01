package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/Enilsonn/CRUD-Postgres/internal/service"
	"github.com/Enilsonn/CRUD-Postgres/internal/utils"
)

type ReportHandler struct {
	service *service.ReportService
}

func NewReportHandler(service *service.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) SellerMonthlySales(w http.ResponseWriter, r *http.Request) {
	var monthPtr *time.Time
	if raw := strings.TrimSpace(r.URL.Query().Get("month")); raw != "" {
		parsed, err := time.Parse("2006-01", raw)
		if err != nil {
			utils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
				"error":   true,
				"code":    "INVALID_MONTH",
				"message": "month must be in YYYY-MM format",
			})
			return
		}
		parsed = parsed.UTC()
		monthPtr = &parsed
	}

	reports, err := h.service.SellerMonthlySales(r.Context(), monthPtr)
	if err != nil {
		utils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error":   true,
			"code":    "REPORT_FAILED",
			"message": err.Error(),
		})
		return
	}

	utils.EncodeJson(w, r, http.StatusOK, reports)
}
