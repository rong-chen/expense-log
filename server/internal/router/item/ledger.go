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

func NewLedgerRouter(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, jwtCfg model.JWTConfig) {
	repo := repository.NewLedgerRepository(db)
	serv := service.NewLedgerService(repo)
	ctrl := controller.NewLedgerController(serv)

	ledgerGroup := r.Group("/ledger")
	ledgerGroup.Use(middleware.JWTAuth([]byte(jwtCfg.Secret), rdb))
	{
		ledgerGroup.POST("", ctrl.CreateSharedLedger)
		ledgerGroup.GET("", ctrl.GetUserLedgers)
		ledgerGroup.POST("/join", ctrl.JoinLedger)
	}
}
