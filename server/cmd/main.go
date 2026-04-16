package main

import (
	"expense-log/internal/configs"
	"expense-log/internal/model"
	"expense-log/internal/router"
	"expense-log/pkg/database"
	"flag"
	"fmt"

	"gorm.io/gorm"
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
		&model.InvitationCode{},
		&model.Tag{},
		&model.Ledger{},
		&model.LedgerMember{},
	); err != nil {
		panic(fmt.Errorf("表结构迁移失败: %w", err))
	}

	// 兼容性扫描：为所有旧用户创建默认个人账本，并将无账本的账单转移过去
	RunBackwardCompatibilityMigration(db)

	rdb := database.InitRedis(conf.Cache.Redis)

	router.Start(db, rdb, conf)
}

func RunBackwardCompatibilityMigration(db *gorm.DB) {
	fmt.Println("Running backward compatibility migration for ledgers...")
	var users []model.User
	db.Find(&users)

	for _, user := range users {
		var ledgerCount int64
		db.Model(&model.Ledger{}).Where("owner_id = ? AND type = ?", user.ID, model.LedgerTypePersonal).Count(&ledgerCount)
		
		if ledgerCount == 0 {
			// 创建默认个人账本
			personalLedger := &model.Ledger{
				Name:    "个人账本",
				OwnerID: user.ID,
				Type:    model.LedgerTypePersonal,
			}
			db.Create(personalLedger)
			
			// 关联 LedgerMember
			db.Create(&model.LedgerMember{
				LedgerID: personalLedger.ID,
				UserID:   user.ID,
				Role:     model.LedgerRoleOwner,
			})
			
			// 将属于该用户但尚未绑定 Ledger 的旧账单更新为属于此个人账本
			db.Model(&model.Bill{}).Where("user_id = ? AND ledger_id IS NULL", user.ID).Update("ledger_id", personalLedger.ID)
			fmt.Printf("Migrated legacy bills for user %v to personal ledger %v\n", user.Email, personalLedger.ID)
		}
	}
}
