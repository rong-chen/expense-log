package item

import (
	"expense-log/internal/controller"
	"expense-log/internal/middleware"
	"expense-log/internal/model"
	"expense-log/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// NewUkeyRouter 注册 Ukey (API Key) 路由
func NewUkeyRouter(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, jwtCfg model.JWTConfig, domain string) {
	ukeyServ := service.NewUkeyService(db, rdb)
	ukeyCtrl := controller.NewUkeyController(ukeyServ, domain)

	// Ukey 的管理挂载在 /user 下
	userGroup := r.Group("/user")
	
	// 需要 JWT 登录才能管理 Ukey
	authGroup := userGroup.Group("")
	authGroup.Use(middleware.JWTAuth([]byte(jwtCfg.Secret)))
	
	authGroup.POST("/ukey", ukeyCtrl.CreateUkey)
	authGroup.GET("/ukey", ukeyCtrl.ListUkeys)
	authGroup.DELETE("/ukey/:id", ukeyCtrl.DeleteUkey)
}
