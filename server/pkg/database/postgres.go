package database

import (
	"expense-log/internal/model"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitPostgres(cfg model.PostgresConfig) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	var err error
	// 1. 建立连接 (禁用外键约束生成，避免历史遗留的数据类型如 varchar与uuid 冲突导致 AutoMigrate 失败)
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}

	// 2. 配置连接池 (从你的 config.yaml 获取参数)
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// 处理存活时间字符串 (例如 "1h")
	duration, _ := time.ParseDuration(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxLifetime(duration)

	return DB
}
