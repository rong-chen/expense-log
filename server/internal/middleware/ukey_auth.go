package middleware

import (
	"expense-log/internal/repository"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UkeyAuth Ukey 自动化凭证鉴权中间件
func UkeyAuth(db *gorm.DB) gin.HandlerFunc {
	ukeyRepo := repository.NewUkeyRepository(db)

	return func(c *gin.Context) {
		// 1. 获取 Authorization header，如果没有则尝试从 URL 参数 ukey 获取
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			authHeader = c.Query("ukey")
		}

		if authHeader == "" {
			c.String(http.StatusUnauthorized, "failed: missing Authorization header or ukey query param")
			c.Abort()
			return
		}

		// 2. 提取 Ukey Secret (支持 Bearer 前缀或纯文本)
		var ukeySecret string
		if strings.HasPrefix(authHeader, "Bearer ") {
			ukeySecret = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			ukeySecret = authHeader
		}

		// 3. 校验 Ukey 是否存在
		ukey, err := ukeyRepo.GetBySecret(ukeySecret)
		if err != nil {
			c.String(http.StatusUnauthorized, "failed: invalid ukey")
			c.Abort()
			return
		}

		// 4. 将提取到的 userID 设置到 Context 中，保持与 JWTAuth 完全一致的行为
		c.Set("userID", ukey.UserID)
		c.Next()
	}
}
