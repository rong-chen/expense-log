package item

import (
	"expense-log/internal/controller"
	"expense-log/internal/middleware"
	"expense-log/internal/model"
	"expense-log/internal/repository"
	"expense-log/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func NewUserRouter(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, jwtCfg model.JWTConfig) {
	r := router.Group("/user")
	// 注册所有层
	repo := repository.NewUserRepository(db)
	invRepo := repository.NewInvitationRepository(db)
	serv := service.NewUserService(repo, invRepo, rdb, jwtCfg)
	con := controller.NewUserController(serv, jwtCfg)

	// 公开路由 (无需认证)
	{
		r.POST("/register", con.Register) // 重新启用邀请码注册
		r.POST("/login", con.Login)
		r.POST("/refresh", con.RefreshToken)
	}

	// 受保护路由 (需 JWT 中间件)
	authGroup := r.Group("")
	authGroup.Use(middleware.JWTAuth([]byte(jwtCfg.Secret), rdb))
	{
		authGroup.GET("/info", con.GetUserInfo)
		authGroup.POST("/logout", con.Logout)
		authGroup.POST("/password", con.UpdatePassword)
	}
}
