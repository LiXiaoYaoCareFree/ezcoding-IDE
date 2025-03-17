package db

import (
	"context"
	"entgo.io/ent/dialect/sql" // ent ORM 的 SQL 驱动
	"fmt"
	"github.com/gofiber/fiber/v2/log" // Fiber 的日志库
	"liteide-backend/config"          // 应用配置管理包
	"liteide-backend/ent"             // ent ORM 生成的数据库客户端
	"time"
)

// InitMySQL 负责初始化 MySQL 连接，并返回 ent 客户端
func InitMySQL(config config.MySQLConfig) *ent.Client {
	var err error

	// 使用 ent 的 SQL 驱动打开 MySQL 连接
	drv, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=True",
		config.Username, // MySQL 用户名
		config.Password, // MySQL 密码
		config.Address,  // MySQL 地址 (host:port)
		config.Database, // MySQL 数据库名称
	))
	if err != nil {
		// 如果连接数据库失败，记录错误日志并终止程序
		log.Fatalf("failed opening connection to mysql: %v", err)
	}

	// 获取底层数据库连接池对象
	db := drv.DB()

	// 设置数据库连接池参数
	db.SetMaxIdleConns(10)           // 设置最大空闲连接数（减少数据库负载）
	db.SetMaxOpenConns(100)          // 设置最大打开连接数（限制并发连接数）
	db.SetConnMaxLifetime(time.Hour) // 设置连接的最大生命周期，防止连接长时间占用

	// 使用 ent ORM 创建并返回一个数据库客户端
	return ent.NewClient(ent.Driver(drv))
}

// WithTx 通过事务执行数据库操作
// - `client` 是数据库客户端
// - `ctx` 是上下文（用于控制超时和取消）
// - `fn` 是要在事务中执行的回调函数
func WithTx(client *ent.Client, ctx context.Context, fn func(client *ent.Client, ctx context.Context) error) error {
	// 开始一个新的事务
	tx, err := client.Tx(ctx)
	if err != nil {
		return err // 如果事务创建失败，直接返回错误
	}

	// 确保在发生 panic 时回滚事务
	defer func() {
		if v := recover(); v != nil { // 捕获 panic
			err := tx.Rollback() // 回滚事务
			if err != nil {
				panic(err) // 如果回滚失败，直接 panic
			}
			panic(v) // 重新抛出 panic
		}
	}()

	// 执行回调函数（事务操作）
	if err := fn(tx.Client(), ctx); err != nil {
		// 如果事务执行过程中发生错误，则回滚
		if res := tx.Rollback(); res != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, res)
		}
		return err
	}

	// 如果事务执行成功，则提交
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}
