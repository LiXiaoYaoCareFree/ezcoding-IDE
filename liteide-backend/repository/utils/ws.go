package utils

import (
	"bufio"
	"context"
	"github.com/gofiber/contrib/websocket" // 引入 Fiber WebSocket 库
	"github.com/gofiber/fiber/v2/log"      // 引入 Fiber 的日志库
	"io"                                   // 引入标准 I/O 库，用于流式读写
	"sync"                                 // 引入 sync 库，用于同步控制
)

// WSWriterCopy 从 `reader` 读取数据，并通过 WebSocket 连接 `writer` 发送数据
// - `reader`：`bufio.Reader` 作为数据源
// - `writer`：`websocket.Conn` 作为数据目标（WebSocket 连接）
// - `wg`：等待组，用于同步多个 goroutine
// - `cancel`：取消函数，触发上下文取消
func WSWriterCopy(reader *bufio.Reader, writer *websocket.Conn, wg *sync.WaitGroup, cancel context.CancelFunc) {
	// 定义缓冲区（1024 字节）
	buf := make([]byte, 1024)

	// 确保在函数结束时调用 `cancel()` 取消操作
	defer cancel()
	// 确保 `wg.Done()` 被调用，减少等待组计数
	defer wg.Done()

	for {
		// 读取 `reader` 中的数据到 `buf`
		nr, err := reader.Read(buf)

		// 如果成功读取了数据
		if nr > 0 {
			// 通过 WebSocket 发送数据（以 BinaryMessage 类型发送）
			err := writer.WriteMessage(websocket.BinaryMessage, buf[0:nr])
			if err != nil {
				return // 发送失败，直接退出
			}
		}

		// 如果读取过程中发生错误，直接退出循环
		if err != nil {
			return
		}
	}
}

// WSReaderCopy 从 WebSocket 连接 `reader` 读取数据，并写入 `writer`（通常是标准输出或文件）
// - `reader`：WebSocket 连接，作为数据源
// - `writer`：`io.Writer` 作为数据目标
// - `wg`：等待组，用于同步多个 goroutine
// - `cancel`：取消函数，触发上下文取消
func WSReaderCopy(reader *websocket.Conn, writer io.Writer, wg *sync.WaitGroup, cancel context.CancelFunc) {
	// 确保在函数结束时取消上下文
	defer cancel()
	// 确保 `wg.Done()` 被调用，减少等待组计数
	defer wg.Done()

	for {
		// 从 WebSocket 连接读取消息
		messageType, p, err := reader.ReadMessage()
		if err != nil {
			// 如果错误不是正常关闭或客户端断开，则记录错误
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) &&
				!websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log.Errorf("failed to read from ws: %v", err)
			}
			return // 发生错误时，退出循环
		}

		// 只处理 TextMessage 类型的消息
		if messageType == websocket.TextMessage {
			_, err := writer.Write(p)
			if err != nil {
				return // 写入失败，直接退出
			}
		}
	}
}
