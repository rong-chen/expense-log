package middleware

import (
	"expense-log/internal/repository"
	"expense-log/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminAuth 管理员权限校验中间件
func AdminAuth(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("userID")
		if !exists {
			response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
			c.Abort()
			return
		}

		userID, ok := userIDValue.(uuid.UUID)
		if !ok {
			response.Fail(c, http.StatusInternalServerError, 50001, "用户ID解析失败")
			c.Abort()
			return
		}

		user, err := userRepo.GetUserByID(userID)
		if err != nil || user.Role != "admin" {
			response.Fail(c, http.StatusForbidden, 40300, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}
