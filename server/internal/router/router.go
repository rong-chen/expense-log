package router

import (
	"expense-log/internal/model"
	"expense-log/internal/router/item"
	"expense-log/pkg/llm"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func Start(db *gorm.DB, rdb *redis.Client, cfg *model.Config) {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	// 默认路径/api
	api := r.Group("/api")
	// 初始版本号
	v1 := api.Group("/v1")

	// 初始化 LLM Provider
	llmProvider, err := llm.New(llm.Config{
		Provider: cfg.LLM.Provider,
		APIKey:   cfg.LLM.APIKey,
		BaseURL:  cfg.LLM.BaseURL,
		Model:    cfg.LLM.Model,
	})
	if err != nil {
		fmt.Printf("⚠️ LLM 初始化失败(图片分析将不可用): %v\n", err)
	}

	// 注册用户路由
	item.NewUserRouter(v1, db, cfg.JWT)

	// 注册 Ukey 自动鉴权路由
	item.NewUkeyRouter(v1, db, rdb, cfg.JWT, cfg.Server.GetDomain())

	// 注册邮箱路由，并返回 emailService
	emailServ := item.NewEmailRouter(v1, db, cfg)

	// 注册账单路由（注入 LLM Provider）
	item.NewBillRouter(v1, db, rdb, cfg.JWT, llmProvider)

	// 注册周期账单路由
	recurringServ := item.NewRecurringBillRouter(v1, db, cfg.JWT)

	// 启动邮件定时拉取后台任务（goroutine，非阻塞）
	interval, err := time.ParseDuration(cfg.Email.PollInterval)
	if err != nil {
		interval = 5 * time.Minute
	}
	emailServ.StartScheduler(interval)

	// 启动周期账单定时调度器（每天凌晨 00:05 自动扣款）
	recurringServ.StartScheduler()

	// 启动 HTTP 服务（阻塞）
	port := strconv.Itoa(cfg.Server.Port)
	fmt.Printf("服务启动中... 监听端口: %s\n", port)

	if err := r.Run(":" + port); err != nil {
		panic(fmt.Errorf("服务启动失败: %w", err))
	}
}
