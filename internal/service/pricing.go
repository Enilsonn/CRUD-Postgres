package service

import (
	"context"
	"math"

	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
)

type PricingService struct {
	Repo *repository.PricingRepository
}

func NewPricingService(repo *repository.PricingRepository) *PricingService {
	return &PricingService{Repo: repo}
}

func (s *PricingService) ComputeCredits(ctx context.Context, model string, pt, ct int64) (int64, float64, float64, error) {
	ppk := 1.0
	cpk := 1.0

	if s != nil && s.Repo != nil {
		if ctx == nil {
			ctx = context.Background()
		}

		matchPPK, matchCPK, ok, err := s.Repo.FindRate(ctx, model)
		if err != nil {
			return 0, 0, 0, err
		}
		if ok {
			ppk = matchPPK
			cpk = matchCPK
		}
	}

	credits := int64(math.Ceil((float64(pt)*ppk + float64(ct)*cpk) / 1000.0))
	return credits, ppk, cpk, nil
}
