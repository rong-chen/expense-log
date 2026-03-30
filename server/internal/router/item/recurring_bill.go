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

// NewRecurringBillRouter 注册周期账单路由，并返回 service 用于启动调度器
func NewRecurringBillRouter(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, jwtCfg model.JWTConfig) service.RecurringBillService {
	serv := service.NewRecurringBillService(db)
	ctrl := controller.NewRecurringBillController(serv)

	recurringGroup := r.Group("/recurring")
	recurringGroup.Use(middleware.JWTAuth([]byte(jwtCfg.Secret), rdb))

	recurringGroup.GET("", ctrl.List)
	recurringGroup.POST("", ctrl.Create)
	recurringGroup.PUT("/:id", ctrl.Update)
	recurringGroup.DELETE("/:id", ctrl.Delete)
	recurringGroup.PATCH("/:id/toggle", ctrl.ToggleActive)

	return serv
}
