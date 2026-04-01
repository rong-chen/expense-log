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

func NewInvitationRouter(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, jwtCfg model.JWTConfig) {
	r := router.Group("/invitation")

	userRepo := repository.NewUserRepository(db)
	invRepo := repository.NewInvitationRepository(db)
	serv := service.NewInvitationService(invRepo)
	con := controller.NewInvitationController(serv)

	r.Use(middleware.JWTAuth([]byte(jwtCfg.Secret), rdb))
	r.Use(middleware.AdminAuth(userRepo))
	{
		r.POST("/generate", con.Generate)
		r.GET("/list", con.List)
	}
}
