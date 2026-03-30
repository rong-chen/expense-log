package middleware

import (
	"context"
	"expense-log/pkg/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 时间窗口内允许的最大请求数
	MaxRequests int
	// 时间窗口大小
	Window time.Duration
}

// RateLimit 基于 Redis 的限流中间件
// 双层策略：
//   - 已登录用户 → 按 userID 限流（不管换多少 IP/代理都精准限制）
//   - 未登录访客 → 按 IP 限流（防暴力登录、注册攻击）
func RateLimit(rdb *redis.Client, jwtSecret []byte, cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// 1. 尝试从 JWT 中解析 userID 作为限流标识
		limitKey := "ip:" + c.ClientIP() // 默认用 IP

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				claims, err := utils.ParseToken(parts[1], jwtSecret)
				if err == nil && claims.UserID.String() != "" {
					// 登录用户：用 userID 作为限流 key
					limitKey = "user:" + claims.UserID.String()
				}
			}
		}

		key := fmt.Sprintf("rate_limit:%s", limitKey)

		// 2. Redis INCR + EXPIRE 固定窗口计数
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// Redis 故障时放行
			c.Next()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, cfg.Window)
		}

		// 3. 设置响应头
		remaining := cfg.MaxRequests - int(count)
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.MaxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		// 4. 超限则拒绝
		if int(count) > cfg.MaxRequests {
			ttl, _ := rdb.TTL(ctx, key).Result()
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": 42900,
				"msg":  fmt.Sprintf("请求过于频繁，请 %d 秒后再试", int(ttl.Seconds())),
			})
			return
		}

		c.Next()
	}
}
