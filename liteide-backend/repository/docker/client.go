package docker

import (
	"github.com/docker/docker/client" // 引入 Docker 客户端库，用于与 Docker 守护进程（daemon）交互
	"github.com/gofiber/fiber/v2/log" // 引入 Fiber 的日志库
)

// InitDocker 初始化并返回 Docker 客户端实例
func InitDocker() *client.Client {
	// 创建 Docker 客户端，使用环境变量加载配置，并启用 API 版本协商
	dockerClient, err := client.NewClientWithOpts(
		client.FromEnv,                     // 从环境变量加载 Docker 配置（如 DOCKER_HOST）
		client.WithAPIVersionNegotiation(), // 自动协商与 Docker 服务器的 API 版本，确保兼容性
	)

	// 如果创建客户端失败，记录错误并终止程序
	if err != nil {
		log.Fatalf("failed to initialize Docker client: %v", err)
	}

	// 返回初始化后的 Docker 客户端
	return dockerClient
}
