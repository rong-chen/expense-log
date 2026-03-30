package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=expense_log_admin password=expense_log_postgres@!! dbname=expense_log port=5433 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database:", err)
		return
	}

	userID := "019d32fd-3127-7956-a635-d478e1413e5d"

	// 物理删除这两笔测试账单，释放唯一指纹锁
	res := db.Exec("DELETE FROM bills WHERE user_id = ? AND amount IN (?, ?)", userID, 15.00, 277.47)
	fmt.Printf("成功从数据库物理删除 %d 笔测试账单\n", res.RowsAffected)

	// 连接 Redis 清除该用户的面板缓存
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6381",
		Password: "expense_log_redis@!!",
		DB:       0,
	})
	
	ctx := context.Background()
	rdb.Del(ctx,
		fmt.Sprintf("user:%s:stats:trend", userID),
		fmt.Sprintf("user:%s:stats:category", userID),
		fmt.Sprintf("user:%s:stats:dashboard", userID),
	)
	fmt.Println("成功重置该用户的所有 Redis 数据面板缓存。")
}
