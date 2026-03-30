package middleware

import (
	"context"
	"expense-log/pkg/response"
	"expense-log/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// JWTAuth JWT 认证中间件 (带 SSO 单点登录校验)
func JWTAuth(secret []byte, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 Authorization Header 提取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, http.StatusUnauthorized, 40100, "缺少认证信息")
			c.Abort()
			return
		}

		// 2. 校验 Bearer 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Fail(c, http.StatusUnauthorized, 40101, "认证格式错误，需要 Bearer Token")
			c.Abort()
			return
		}

		// 3. 解析 Token
		claims, err := utils.ParseToken(parts[1], secret)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, 40102, "Token 无效或已过期")
			c.Abort()
			return
		}

		// 4. 校验必须是 Access Token
		if claims.TokenType != utils.TokenTypeAccess {
			response.Fail(c, http.StatusUnauthorized, 40103, "请使用 Access Token 进行认证")
			c.Abort()
			return
		}

		// 4.1. 单点登录校验 (如果 token 中包含 sessionID)
		if claims.SessionID != "" && rdb != nil {
			activeSession, err := rdb.Get(context.Background(), "session:"+claims.UserID.String()).Result()
			if err != nil || activeSession != claims.SessionID {
				// redis 中没找到 session 或者与当前不匹配 -> 说明其他设备重新登录了
				response.Fail(c, http.StatusUnauthorized, 40106, "您的账号已在其他设备登录，请重新登录")
				c.Abort()
				return
			}
		}

		// 5. 将 userID 存入上下文
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
