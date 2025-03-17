package config

import "liteide-backend/repository/utils" // 引入工具包，用于解析环境变量

// ApiConfig 结构体定义 API 相关配置
type ApiConfig struct {
	Port int // API 监听端口
}

// MySQLConfig 结构体定义 MySQL 数据库的连接配置
type MySQLConfig struct {
	Username string // MySQL 用户名
	Password string // MySQL 密码
	Address  string // MySQL 服务器地址 (host:port)
	Database string // MySQL 数据库名称
}

// AppConfig 结构体定义整个应用的配置信息
type AppConfig struct {
	ApiConfig              ApiConfig   // API 配置
	MySQLConfig            MySQLConfig // MySQL 连接配置
	ContainerServicePrefix string      // 容器服务前缀（用于 Swarm 容器命名）
	DataDirectory          string      // 应用数据存储目录
}

// NewConfig 创建并返回应用的默认配置
// - 通过 `utils.ParseEnvConfig` 解析环境变量（如果未设置，则使用默认值）
func NewConfig() AppConfig {
	return AppConfig{
		// 解析 API 端口，默认为 8080
		ApiConfig: ApiConfig{
			Port: utils.ParseEnvConfig("WEB_PORT", 8080),
		},
		// 解析 MySQL 连接参数
		MySQLConfig: MySQLConfig{
			Username: utils.ParseEnvConfig("MYSQL_USERNAME", "root"),       // 解析 MySQL 用户名，默认 root
			Password: utils.ParseEnvConfig("MYSQL_PASSWORD", "123456"),     // 解析 MySQL 密码，默认 123456
			Address:  utils.ParseEnvConfig("MYSQL_ADDR", "localhost:3306"), // 解析 MySQL 地址，默认 localhost:3306
			Database: utils.ParseEnvConfig("MYSQL_DATABASE", "liteide"),    // 解析 MySQL 数据库名称，默认 liteide
		},
		ContainerServicePrefix: "liteide-pod-",                                  // Swarm 容器服务的命名前缀
		DataDirectory:          "D:/Proj/ezcoding/liteide/liteide-backend/data", // 存储数据的本地目录
	}
}
