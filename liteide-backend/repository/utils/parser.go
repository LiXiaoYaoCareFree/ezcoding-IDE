package utils

import (
	"github.com/gofiber/fiber/v2/log" // 引入 Fiber 的日志库
	"net"                             // 用于网络相关操作
	"os"                              // 用于获取环境变量
	"strconv"                         // 用于字符串转换
)

// ParseEnvConfig 解析环境变量，并根据提供的默认值类型进行转换
// - `key`：环境变量的键名
// - `defaultValue`：默认值，如果环境变量不存在，则返回这个值
// - `T`：泛型，支持 string、int、bool 三种类型
func ParseEnvConfig[T string | int | bool](key string, defaultValue T) T {
	// 获取环境变量的值
	value, exists := os.LookupEnv(key)
	if !exists {
		// 如果环境变量不存在，则返回默认值
		return defaultValue
	}

	// 根据 defaultValue 的类型转换环境变量的值
	switch any(defaultValue).(type) {
	case string:
		// 如果默认值是 string，则直接返回环境变量的值
		return any(value).(T)
	case int:
		// 如果默认值是 int，则尝试将环境变量转换为整数
		intValue, err := strconv.Atoi(value)
		if err != nil {
			log.Panic(err) // 解析失败时，终止程序并打印错误
		}
		return any(intValue).(T)
	case bool:
		// 如果默认值是 bool，则尝试将环境变量转换为布尔值
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			log.Panic(err) // 解析失败时，终止程序并打印错误
		}
		return any(boolValue).(T)
	default:
		// 如果 defaultValue 不是受支持的类型，则触发 panic
		panic("unknown type of defaultValue")
	}
}

// GetPortFromAddress 从 `host:port` 格式的地址中提取端口
// - `addr`：网络地址字符串，例如 "127.0.0.1:8080"
// - 返回端口字符串，例如 "8080"
func GetPortFromAddress(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatal(err) // 解析失败时终止程序并打印错误
	}
	return port
}
