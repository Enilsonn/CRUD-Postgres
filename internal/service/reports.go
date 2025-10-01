package service

import (
	"context"
	"time"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
)

type ReportService struct {
	reports *repository.ReportRepository
}

func NewReportService(reports *repository.ReportRepository) *ReportService {
	return &ReportService{reports: reports}
}

func (s *ReportService) SellerMonthlySales(ctx context.Context, month *time.Time) ([]model.SellerMonthlySales, error) {
	return s.reports.ListSellerMonthlySales(ctx, month)
}
