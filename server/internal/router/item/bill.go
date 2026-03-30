package item

import (
	"expense-log/internal/controller"
	"expense-log/internal/middleware"
	"expense-log/internal/model"
	"expense-log/internal/service"
	"expense-log/pkg/llm"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// NewBillRouter 注册账单路由
func NewBillRouter(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, jwtCfg model.JWTConfig, llmProvider llm.Provider) {
	billServ := service.NewBillService(db, rdb)
	billCtrl := controller.NewBillController(billServ, db, llmProvider)

	billGroup := r.Group("/bill")

	// iOS 快捷指令专用：使用 Ukey 鉴权保护
	ukeyGroup := billGroup.Group("")
	ukeyGroup.Use(middleware.UkeyAuth(db))
	ukeyGroup.POST("/image", billCtrl.UploadImageReceipt)

	// 正常的 Web 端接口需要走 JWT 拦截
	webGroup := billGroup.Group("")
	webGroup.Use(middleware.JWTAuth([]byte(jwtCfg.Secret), rdb))

	webGroup.GET("/stats/trend", billCtrl.GetTrendStats)
	webGroup.GET("/stats/category", billCtrl.GetCategoryStats)
	webGroup.GET("/dashboard", billCtrl.GetDashboardStats)
	webGroup.GET("/list", billCtrl.GetBillList)
	webGroup.GET("/:id", billCtrl.GetBillDetail)
	
	// 为 Web App 开放独立的图片识别网关，防止和 iOS 快捷指令路由冲突
	webGroup.POST("/upload", billCtrl.UploadImageReceipt)
	// 允许手动修改备注
	webGroup.PUT("/:id/remark", billCtrl.UpdateRemark)
	// 允许全量修改及删除
	webGroup.PUT("/:id", billCtrl.UpdateBill)
	webGroup.DELETE("/:id", billCtrl.DeleteBill)
	webGroup.POST("/manual", billCtrl.CreateBill)
}
