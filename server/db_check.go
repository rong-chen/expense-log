package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

type Bill struct {
	ID              string
	UserID          string
	Amount          float64
	Category        string
	Remark          string
	TransactionDate time.Time
	CreatedAt       time.Time
}

func main() {
	dsn := "host=localhost user=expense_log_admin password=expense_log_postgres@!! dbname=expense_log port=5433 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database:", err)
		return
	}

	var bills []Bill
	result := db.Order("created_at desc").Limit(3).Find(&bills)
	if result.Error != nil {
		fmt.Println("Query error:", result.Error)
		return
	}

	fmt.Println("=== 最近录入的账单 (最新 3 条) ===")
	for i, b := range bills {
		fmt.Printf("%d: [UserID: %d] [%s] %.2f - %s (录入时间: %v)\n", i+1, b.UserID, b.Category, b.Amount, b.Remark, b.CreatedAt.Format("2006-01-02 15:04:05"))
	}
}
