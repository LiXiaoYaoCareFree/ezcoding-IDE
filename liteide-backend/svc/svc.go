package svc

import (
	dockerClient "github.com/docker/docker/client" // 引入 Docker 客户端库，用于与 Docker 进行交互
	"liteide-backend/config"                       // 引入配置管理包，用于加载应用配置
	"liteide-backend/ent"                          // 引入 ent ORM 库，用于与数据库交互
	"liteide-backend/repository/db"                // 引入数据库操作包，包含数据库初始化和迁移等功能
	"liteide-backend/repository/docker"            // 引入 Docker 操作包，提供与 Docker 客户端交互的功能
)

// 全局变量 SVC，用于存储应用的 ServiceContext 实例
var SVC *ServiceContext

// ServiceContext 结构体用于存储应用程序所需的所有服务和配置
type ServiceContext struct {
	AppConfig config.AppConfig     // 存储应用程序的配置
	Database  *ent.Client          // 数据库客户端，用于数据库操作
	Docker    *dockerClient.Client // Docker 客户端，用于与 Docker 交互
}

// NewServiceContext 用于初始化 ServiceContext 并将其赋值给全局变量 SVC
func NewServiceContext() {
	// 获取应用的配置
	appConf := config.NewConfig()

	// 初始化 ServiceContext，并将其赋值给全局变量 SVC
	SVC = &ServiceContext{
		AppConfig: appConf,                           // 将应用配置赋值给 ServiceContext
		Database:  db.InitMySQL(appConf.MySQLConfig), // 初始化数据库连接，使用配置中的 MySQL 配置
		Docker:    docker.InitDocker(),               // 初始化 Docker 客户端
	}
}
