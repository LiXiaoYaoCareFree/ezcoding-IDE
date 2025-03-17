package db

import (
	"context"
	"fmt"
	"liteide-backend/config" // 引入配置管理包，用于获取 MySQL 连接信息
	"liteide-backend/ent"    // 引入 ent ORM，用于数据库操作
	"log"                    // 标准日志库，用于输出日志

	_ "github.com/go-sql-driver/mysql" // 引入 MySQL 驱动（仅导入，不直接使用）
)

// Migrate 连接 MySQL 并执行数据库迁移
func Migrate(config config.AppConfig) {
	// 使用 ent.Open 连接 MySQL 数据库
	client, err := ent.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=True",
			config.MySQLConfig.Username, // 数据库用户名
			config.MySQLConfig.Password, // 数据库密码
			config.MySQLConfig.Address,  // 数据库地址 (host:port)
			config.MySQLConfig.Database, // 数据库名称
		))
	if err != nil {
		// 如果连接数据库失败，打印错误日志并终止程序
		log.Panicf("failed opening connection to mysql: %v", err)
	}

	// 确保在函数执行完后关闭数据库连接
	defer func(client *ent.Client) {
		err := client.Close()
		if err != nil {
			// 如果关闭数据库连接失败，记录错误日志
			log.Panicf("failed to close connection: %v", err)
		}
	}(client)

	// 执行数据库模式迁移，确保数据库表结构与 `ent` 定义的模式一致
	if err := client.Schema.Create(context.Background()); err != nil {
		// 如果迁移失败，打印错误日志并终止程序
		log.Panicf("failed creating schema resources: %v", err)
	}
}
