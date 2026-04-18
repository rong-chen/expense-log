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

func NewAdminRouter(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, jwtCfg model.JWTConfig) {
	admin := router.Group("/admin")

	userRepo := repository.NewUserRepository(db)
	billRepo := repository.NewBillRepository(db)
	emailRepo := repository.NewEmailRepository(db)

	serv := service.NewAdminService(userRepo, billRepo, emailRepo)
	con := controller.NewAdminController(serv)

	admin.Use(middleware.JWTAuth([]byte(jwtCfg.Secret), rdb))
	admin.Use(middleware.AdminAuth(userRepo))
	{
		admin.GET("/users", con.ListUsers)
		admin.POST("/role", con.UpdateUserRole)
		admin.POST("/user/password", con.ResetUserPassword)
	}
}
