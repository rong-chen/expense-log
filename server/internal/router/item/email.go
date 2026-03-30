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

func NewEmailRouter(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, cfg *model.Config) service.EmailService {
	r := router.Group("/email")

	emailRepo := repository.NewEmailRepository(db)
	emailServ := service.NewEmailService(emailRepo, cfg.Email)
	emailCon := controller.NewEmailController(emailServ)

	// 所有邮箱接口都需要认证
	r.Use(middleware.JWTAuth([]byte(cfg.JWT.Secret), rdb))
	{
		r.POST("/bind", emailCon.BindEmail)       // 绑定邮箱
		r.GET("/accounts", emailCon.GetEmails)     // 获取绑定列表
		r.DELETE("/:id", emailCon.DeleteEmail)     // 解绑邮箱
	}

	// 返回 emailService 以便 main.go 启动定时任务
	return emailServ
}
