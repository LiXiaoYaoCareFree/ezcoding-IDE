package model

import (
	"github.com/gofiber/fiber/v2/log" // 引入 Fiber 的日志库，用于记录错误信息
	"liteide-backend/ent/property"    // 引入 ent ORM 生成的 property 模型
)

// Language 定义了一种自定义类型，用于表示编程语言
type Language string

// 定义可支持的编程语言常量
const (
	LanguageC      Language = "C"      // C 语言
	LanguagePython Language = "PYTHON" // Python 语言
)

// ToEnt 将自定义的 Language 类型转换为 ent ORM 识别的 property.Language 类型
func (language Language) ToEnt() property.Language {
	switch language {
	case LanguageC:
		return property.LanguageC // 如果是 "C"，返回 ent ORM 对应的值
	case LanguagePython:
		return property.LanguagePython // 如果是 "PYTHON"，返回 ent ORM 对应的值
	default:
		// 如果遇到未知语言，记录错误日志
		log.Errorf("unknown language: %s", language)
		return "" // 返回空字符串，表示转换失败
	}
}
