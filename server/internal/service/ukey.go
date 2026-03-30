package service

import (
	"context"
	"errors"
	"expense-log/internal/model"
	"expense-log/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// UkeyResponse 用于列表返回时的脱敏结构
type UkeyResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	SecretKey string    `json:"secret_key"` // 明文返回
	CreatedAt time.Time `json:"created_at"`
}

type UkeyService interface {
	CreateUkey(userID uuid.UUID, name string) (string, error)
	ListUkeys(userID uuid.UUID) ([]UkeyResponse, error)
	VerifyUkey(secret string) (uuid.UUID, error)
	DeleteUkey(id uuid.UUID, userID uuid.UUID) error
}

type ukeyService struct {
	repo repository.UkeyRepository
	rdb  *redis.Client
}

func NewUkeyService(db *gorm.DB, rdb *redis.Client) UkeyService {
	return &ukeyService{
		repo: repository.NewUkeyRepository(db),
		rdb:  rdb,
	}
}

func (s *ukeyService) CreateUkey(userID uuid.UUID, name string) (string, error) {
	// 业务限制：每人最多只能拥有 1 个激活的 Ukey
	existing, err := s.repo.GetByUserID(userID)
	if err != nil {
		return "", err
	}
	if len(existing) >= 1 {
		return "", errors.New("每个人最多只能创建1个有效凭证，请先删除旧凭证")
	}

	// 生成新的 API Key
	secret := fmt.Sprintf("exp_sk_%s", uuid.New().String())

	ukey := &model.Ukey{
		UserID:    userID,
		SecretKey: secret,
		Name:      name,
	}

	if err := s.repo.Create(ukey); err != nil {
		return "", err
	}
	return secret, nil // 仅有这一次返回明文！
}

func (s *ukeyService) ListUkeys(userID uuid.UUID) ([]UkeyResponse, error) {
	ukeys, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	var res []UkeyResponse
	for _, k := range ukeys {
		res = append(res, UkeyResponse{
			ID:        k.ID,
			Name:      k.Name,
			SecretKey: k.SecretKey,
			CreatedAt: k.CreatedAt,
		})
	}
	return res, nil
}

func (s *ukeyService) VerifyUkey(secret string) (uuid.UUID, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("ukey:%s", secret)

	// 1. 尝试读 Redis 缓存
	userIDStr, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil && userIDStr != "" {
		if id, parseErr := uuid.Parse(userIDStr); parseErr == nil {
			return id, nil
		}
	}

	// 2. 缓存击穿，查数据库
	ukey, err := s.repo.GetBySecret(secret)
	if err != nil {
		return uuid.Nil, errors.New("无效的自动化凭证")
	}

	// 3. 写入缓存回源 (24小时过期)
	s.rdb.Set(ctx, cacheKey, ukey.UserID.String(), 24*time.Hour)

	return ukey.UserID, nil
}

func (s *ukeyService) DeleteUkey(id uuid.UUID, userID uuid.UUID) error {
	// 必须要先查出 Secret 才能清缓存
	ukey, err := s.repo.GetByID(id)
	if err == nil && ukey.UserID == userID {
		// 删除缓存
		s.rdb.Del(context.Background(), fmt.Sprintf("ukey:%s", ukey.SecretKey))
	}

	// 从数据库删除
	return s.repo.Delete(id, userID)
}
