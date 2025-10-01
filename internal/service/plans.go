package service

import (
	"context"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
)

type PlanService struct {
	plans *repository.ProductRepository
}

type PlanSearchFilters struct {
	Name               string
	MinPriceCents      *int64
	MaxPriceCents      *int64
	Category           string
	ManufacturedInMari *bool
}

func NewPlanService(plans *repository.ProductRepository) *PlanService {
	return &PlanService{plans: plans}
}

func (s *PlanService) Search(ctx context.Context, filters PlanSearchFilters) ([]model.Plan, error) {
	repoFilters := repository.PlanSearchFilters{
		Name:               filters.Name,
		MinPriceCents:      filters.MinPriceCents,
		MaxPriceCents:      filters.MaxPriceCents,
		Category:           filters.Category,
		ManufacturedInMari: filters.ManufacturedInMari,
	}
	return s.plans.SearchPlans(ctx, repoFilters)
}

func (s *PlanService) LowStock(ctx context.Context) ([]model.Plan, error) {
	return s.plans.ListLowStock(ctx)
}
