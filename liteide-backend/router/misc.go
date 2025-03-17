package router

import (
	"errors"                               // 标准错误处理包
	"github.com/gofiber/contrib/websocket" // 引入 Fiber 的 WebSocket 库
	"github.com/gofiber/fiber/v2"          // 引入 Fiber Web 框架
	"strconv"                              // 用于字符串转换，如分页参数解析
)

// ErrorHandler 统一错误处理函数
// - `c`：Fiber 上下文对象
// - `err`：发生的错误
func ErrorHandler(c *fiber.Ctx, err error) error {
	// 默认返回 500 服务器内部错误
	code := fiber.StatusInternalServerError

	// 如果错误是 *fiber.Error 类型，则获取其错误码
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	// 返回 JSON 格式的错误信息
	return c.Status(code).JSON(fiber.Map{
		"message": err.Error(), // 返回错误信息
	})
}

// useWS WebSocket 连接检查中间件
// - `c`：Fiber 上下文对象
func useWS(c *fiber.Ctx) error {
	// 检查是否为 WebSocket 连接请求
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next() // 允许继续处理 WebSocket 连接
	}
	// 如果不是 WebSocket 连接，则返回 426 Upgrade Required
	return fiber.ErrUpgradeRequired
}

// usePagination 分页参数解析中间件
// 作用：解析 `page` 和 `size` 参数，并计算 `offset` 和 `limit` 值
func usePagination() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 解析 `page` 参数，默认值为 1
		page, err := strconv.ParseInt(c.Query("page", "1"), 10, 64)
		if page <= 0 || err != nil {
			page = 1 // 如果 `page` 小于等于 0 或解析失败，则默认为 1
		}

		// 解析 `size` 参数，默认值为 10
		size, err := strconv.ParseInt(c.Query("size", "10"), 10, 64)
		if size < 0 || err != nil {
			size = 10 // 如果 `size` 小于 0 或解析失败，则默认为 10
		}

		// 计算偏移量（offset = (page - 1) * size）
		c.Locals("offset", int((page-1)*size))
		// 设置分页限制值
		c.Locals("limit", int(size))

		// 继续执行下一个中间件或路由处理
		return c.Next()
	}
}
