package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type QuotaService interface {
	CheckAndConsumeQuota(ctx context.Context, userID uuid.UUID, consumeCount int, maxQuota int) (int, error)
}

type quotaService struct {
	rdb *redis.Client
}

func NewQuotaService(rdb *redis.Client) QuotaService {
	return &quotaService{
		rdb: rdb,
	}
}

// CheckAndConsumeQuota 检查基于每天的 AI 识图额度并进行扣减
func (s *quotaService) CheckAndConsumeQuota(ctx context.Context, userID uuid.UUID, consumeCount int, maxQuota int) (int, error) {
	if maxQuota <= 0 {
		maxQuota = 10 // 默认保底 10 次
	}

	today := time.Now().Format("2006-01-02")
	cacheKey := fmt.Sprintf("quota:ai_image:%s:%s", userID.String(), today)

	// 使用 Lua 脚本保证原子性判断和扣减
	script := `
		local current = redis.call("GET", KEYS[1])
		if not current then
			current = 0
		else
			current = tonumber(current)
		end
		
		local requested = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		
		if current + requested > limit then
			return -1 -- 额度不足
		end
		
		local new_val = redis.call("INCRBY", KEYS[1], requested)
		-- 如果是新建的键，设置在明天零点左右过期 (24小时)
		if current == 0 then
			redis.call("EXPIRE", KEYS[1], 86400)
		end
		
		return limit - new_val
	`

	result, err := s.rdb.Eval(ctx, script, []string{cacheKey}, consumeCount, maxQuota).Result()
	if err != nil {
		return 0, fmt.Errorf("redis execution failed: %v", err)
	}

	remain := result.(int64)
	if remain == -1 {
		return 0, errors.New("超出了每日可用配额")
	}

	return int(remain), nil
}
