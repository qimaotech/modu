package i18n

import (
	"fmt"
	"os"
	"strings"
)

// Lang 语言类型
type Lang string

const (
	LangZh Lang = "zh"
	LangEn Lang = "en"
)

// Messages 翻译消息结构
type Messages map[string]string

var (
	// currentLang 当前语言
	currentLang Lang = LangZh
	// messages 翻译消息映射
	messages = make(map[Lang]Messages)
)

// Init 初始化 i18n
func Init() {
	// 默认中文，只在明确设置为英文时才切换
	lang := os.Getenv("LANG")
	if strings.HasPrefix(strings.ToLower(lang), "en") {
		currentLang = LangEn
	}
	messages[LangZh] = zhMessages
	messages[LangEn] = enMessages
}

// SetLang 设置语言
func SetLang(lang Lang) {
	currentLang = lang
}

// T 获取翻译消息
func T(key string) string {
	if msg, ok := messages[currentLang][key]; ok {
		return msg
	}
	// 如果当前语言没有，回退到中文
	if msg, ok := messages[LangZh][key]; ok {
		return msg
	}
	return key
}

// Tprintf 格式化翻译消息
func Tprintf(key string, args ...interface{}) string {
	return fmt.Sprintf(T(key), args...)
}

// MustGet 获取翻译消息，如果不存在返回默认值
func MustGet(key string, defaultMsg string) string {
	if msg, ok := messages[currentLang][key]; ok {
		return msg
	}
	return defaultMsg
}

func init() {
	Init()
}
