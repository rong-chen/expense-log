package service

import (
	"crypto/rand"
	"expense-log/internal/model"
	"expense-log/internal/repository"
	"math/big"
)

type InvitationService interface {
	Generate(count int, role string) ([]string, error)
	List() ([]model.InvitationCode, error)
}

type invitationService struct {
	repo repository.InvitationRepository
}

func NewInvitationService(repo repository.InvitationRepository) InvitationService {
	return &invitationService{
		repo: repo,
	}
}

func (s *invitationService) Generate(count int, role string) ([]string, error) {
	var codes []string
	for i := 0; i < count; i++ {
		code, err := generateRandomCode(8)
		if err != nil {
			return nil, err
		}
		invitation := &model.InvitationCode{
			Code:   code,
			Role:   role,
			IsUsed: false,
		}
		if err := s.repo.Create(invitation); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, nil
}

func (s *invitationService) List() ([]model.InvitationCode, error) {
	return s.repo.ListAll()
}

func generateRandomCode(length int) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 排除容易混淆的字符 I, O, 0, 1
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}
