package item

import (
	"expense-log/internal/controller"
	"expense-log/internal/middleware"
	"expense-log/internal/model"
	"expense-log/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// NewTagRouter 注册标签路由
func NewTagRouter(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, jwtCfg model.JWTConfig) {
	tagRepo := repository.NewTagRepository(db)
	tagCtrl := controller.NewTagController(tagRepo)

	tagGroup := r.Group("/tag")
	tagGroup.Use(middleware.JWTAuth([]byte(jwtCfg.Secret), rdb))

	tagGroup.GET("/list", tagCtrl.ListTags)
	tagGroup.POST("/create", tagCtrl.CreateTag)
	tagGroup.DELETE("/:id", tagCtrl.DeleteTag)

	// 账单-标签关联
	tagGroup.POST("/bill/:id", tagCtrl.SetBillTags)
	tagGroup.GET("/bill/:id", tagCtrl.GetBillTags)
}
