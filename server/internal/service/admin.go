package service

import (
	"expense-log/internal/model"
	"expense-log/internal/repository"

	"github.com/google/uuid"
)

type AdminService interface {
	ListUsers(page, pageSize int) ([]model.User, int64, error)
	UpdateUserRole(id uuid.UUID, role string) error
	GetSystemStats() (map[string]int64, error)
}

type adminService struct {
	userRepo          repository.UserRepository
	billRepo          repository.BillRepository
	emailRepo         repository.EmailRepository
	recurringBillRepo repository.RecurringBillRepository
}

func NewAdminService(
	userRepo repository.UserRepository,
	billRepo repository.BillRepository,
	emailRepo repository.EmailRepository,
	recurringBillRepo repository.RecurringBillRepository,
) AdminService {
	return &adminService{
		userRepo:          userRepo,
		billRepo:          billRepo,
		emailRepo:         emailRepo,
		recurringBillRepo: recurringBillRepo,
	}
}

func (s *adminService) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	return s.userRepo.ListAll(page, pageSize)
}

func (s *adminService) UpdateUserRole(id uuid.UUID, role string) error {
	return s.userRepo.UpdateUserRole(id, role)
}

func (s *adminService) GetSystemStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	userCount, _ := s.userRepo.CountTotal()
	billCount, _ := s.billRepo.CountTotal()
	emailAccountCount, _ := s.emailRepo.CountTotalAccounts()
	recurringCount, _ := s.recurringBillRepo.CountTotal()

	stats["total_users"] = userCount
	stats["total_bills"] = billCount
	stats["total_email_accounts"] = emailAccountCount
	stats["total_recurring_tasks"] = recurringCount

	return stats, nil
}
