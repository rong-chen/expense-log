package item

import (
	"expense-log/internal/controller"
	"expense-log/internal/middleware"
	"expense-log/internal/model"
	"expense-log/internal/repository"
	"expense-log/pkg/llm"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// NewMCPRouter 注册 Model Context Protocol (MCP) 路由端点
func NewMCPRouter(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, jwtCfg model.JWTConfig, domain string, llmProvider llm.Provider) {
	mcpCtrl := controller.NewMCPController(db, domain, llmProvider)
	userRepo := repository.NewUserRepository(db)

	// 1. 公共 MCP 标准协议信道 (基于 SSE 传输协议)
	// 使用时请求参数形式携带 ?api_key=xxx
	r.GET("/mcp/sse", mcpCtrl.HandleSSE)
	r.POST("/mcp/message", mcpCtrl.HandleMessage)

	// 2. 普通用户申请/查看 API Key 路由
	userGroup := r.Group("/user")
	authGroup := userGroup.Group("")
	authGroup.Use(middleware.JWTAuth([]byte(jwtCfg.Secret), rdb))
	{
		authGroup.POST("/mcp/keys/apply", mcpCtrl.ApplyMCPKey)
		authGroup.GET("/mcp/keys/my", mcpCtrl.ListMyMCPKeys)
	}

	// 3. 管理后台审批 API Key 路由 (必须为管理员角色)
	adminGroup := r.Group("/admin")
	adminAuthGroup := adminGroup.Group("")
	adminAuthGroup.Use(middleware.JWTAuth([]byte(jwtCfg.Secret), rdb))
	adminAuthGroup.Use(middleware.AdminAuth(userRepo))
	{
		adminAuthGroup.GET("/mcp/keys", mcpCtrl.AdminListMCPKeys)
		adminAuthGroup.POST("/mcp/keys/:id/approve", mcpCtrl.AdminApproveMCPKey)
		adminAuthGroup.POST("/mcp/keys/:id/reject", mcpCtrl.AdminRejectMCPKey)
		adminAuthGroup.POST("/mcp/keys/:id/revoke", mcpCtrl.AdminRevokeMCPKey)
	}
}
