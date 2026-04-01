package service

import (
	"expense-log/internal/model"
	"expense-log/internal/repository"

	"github.com/google/uuid"
)

type AdminService interface {
	ListUsers(page, pageSize int) ([]model.User, int64, error)
	UpdateUserRole(id uuid.UUID, role string) error
}

type adminService struct {
	userRepo  repository.UserRepository
	billRepo  repository.BillRepository
	emailRepo repository.EmailRepository
}

func NewAdminService(
	userRepo repository.UserRepository,
	billRepo repository.BillRepository,
	emailRepo repository.EmailRepository,
) AdminService {
	return &adminService{
		userRepo:  userRepo,
		billRepo:  billRepo,
		emailRepo: emailRepo,
	}
}

func (s *adminService) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	return s.userRepo.ListAll(page, pageSize)
}

func (s *adminService) UpdateUserRole(id uuid.UUID, role string) error {
	return s.userRepo.UpdateUserRole(id, role)
}

