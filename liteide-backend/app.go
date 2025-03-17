package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"liteide-backend/repository/db"
	"liteide-backend/router"
	"liteide-backend/svc"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var ApiServer *fiber.App // 声明一个全局变量，用于存储 Fiber 应用实例

// startApiServer 启动 API 服务器
func startApiServer() {
	// 创建一个新的 Fiber 应用，并配置自定义的错误处理函数
	ApiServer = fiber.New(fiber.Config{
		ErrorHandler: router.ErrorHandler, // 自定义错误处理函数
	})

	// 使用自定义的路由配置
	router.UseRouter(ApiServer)

	// 输出服务器启动的日志，显示监听的端口号
	log.Infof("Web Server listening at %v", svc.SVC.AppConfig.ApiConfig.Port)

	// 启动服务器并监听指定的端口号，若出错则输出日志并终止程序
	if err := ApiServer.Listen(fmt.Sprintf(":%d", svc.SVC.AppConfig.ApiConfig.Port)); err != nil {
		log.Panic(err) // 如果服务器启动失败，输出错误并终止程序
	}
}

func main() {
	// 初始化服务上下文，可能用于加载配置等初始化操作
	svc.NewServiceContext()

	// 执行数据库迁移操作，确保数据库结构与应用一致
	db.Migrate(svc.SVC.AppConfig)

	// 使用 goroutine 异步启动 API 服务器
	go startApiServer()

	// 创建一个信号通道，用于接收操作系统发送的信号（如关闭信号）
	quit := make(chan os.Signal)
	// 注册接收 SIGINT (Ctrl+C) 和 SIGTERM (终止信号)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞直到接收到退出信号
	<-quit

	// 收到退出信号后，记录关闭服务器的日志
	log.Info("Shutdown Server ...")

	// 创建一个带有 2 秒超时的上下文，用于优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel() // 在函数退出时，确保调用 cancel() 释放资源

	// 调用 Fiber 的 Shutdown 方法优雅关闭服务器
	_ = ApiServer.Shutdown()

	// 等待上下文完成（即等待服务器完全关闭）
	<-ctx.Done()

	// 服务器关闭后，记录退出日志
	log.Info("Server quit.")
}
