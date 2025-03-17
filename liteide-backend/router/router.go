package router

import (
	"github.com/gofiber/contrib/websocket" // 引入 Fiber WebSocket 支持库
	"github.com/gofiber/fiber/v2"          // 引入 Fiber Web 框架
	"liteide-backend/controller"           // 引入控制器，用于处理 HTTP 请求
)

// UseRouter 负责注册 API 和 WebSocket 相关的路由
// - `app`：Fiber Web 服务器实例
func UseRouter(app *fiber.App) {
	// 注册 HTTP API 路由
	app.Post("/container", controller.CreateContainer)
	// 处理创建容器请求（POST 方法）
	// 例如：POST /container
	// Body: {"image": "nginx", "name": "my-container"}
	// 返回：{"id": "abc123456", "status": "created"}

	app.Delete("/container/:id<int>", controller.RemoveContainer)
	// 处理删除指定 ID 容器的请求（DELETE 方法）
	// 例如：DELETE /container/123
	// 返回：{"id": 123, "status": "removed"}

	// WebSocket 相关路由
	app.Use("/ws", useWS)
	// 中间件，针对所有 `/ws` 开头的 WebSocket 路由执行额外逻辑（如身份验证）
	// 例如：拦截非授权用户或日志记录

	app.Get("/ws/container/:id<int>", websocket.New(controller.AttachContainer))
	// WebSocket 连接到指定 ID 的 Docker 容器（GET 方法）
	// 例如：ws://localhost:8080/ws/container/123
	// 用于获取容器的实时日志或交互式终端
}
