package service

import (
	"context"
	"fmt"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
	"github.com/Bengkelin/bengkelin-service/internal/pkg/repository"
	"github.com/Bengkelin/bengkelin-service/pkg/helpers"
	applog "github.com/Bengkelin/bengkelin-service/pkg/log"
)

type AdminFeeServiceImpl struct {
	adminFeeRepo repository.AdminFeeRepositoryInterface
}

func NewAdminFeeService(deps ServiceDependencies) AdminFeeServiceInterface {
	return &AdminFeeServiceImpl{
		adminFeeRepo: deps.AdminFeeRepo,
	}
}

func (s *AdminFeeServiceImpl) CreateAdminFee(ctx context.Context, adminFee float64) (*models.AdminFee, error) {
	fee := models.AdminFee{
		ID:       helpers.GenerateUUID(),
		AdminFee: adminFee,
	}

	created, err := s.adminFeeRepo.CreateAdminFee(ctx, fee)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin fee: %w", err)
	}

	applog.InfoCtx(ctx, "Admin fee created", "fee_id", created.ID, "amount", adminFee)
	return &created, nil
}

func (s *AdminFeeServiceImpl) GetLatestAdminFee(ctx context.Context) (*models.AdminFee, error) {
	fee, err := s.adminFeeRepo.GetOneAdminFeeLatest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin fee: %w", err)
	}
	return fee, nil
}
