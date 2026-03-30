package main

import (
	"expense-log/internal/configs"
	"expense-log/internal/model"
	"expense-log/internal/router"
	"expense-log/pkg/database"
	"flag"
	"fmt"
)

func main() {
	configPtr := flag.String("conf", "./configs/config.yaml", "配置文件路径")
	if !flag.Parsed() {
		flag.Parse()
	}
	conf := configs.InitConfigs(*configPtr)
	db := database.InitPostgres(conf.Database.Postgres)

	// AutoMigrate 自动同步数据表结构
	if err := db.AutoMigrate(
		&model.User{},
		&model.EmailRecord{},
		&model.EmailAttachment{},
		&model.UserEmailAccount{},
		&model.Bill{},
		&model.Ukey{},
		&model.RecurringBill{},
	); err != nil {
		panic(fmt.Errorf("表结构迁移失败: %w", err))
	}

	rdb := database.InitRedis(conf.Cache.Redis)

	router.Start(db, rdb, conf)
}
