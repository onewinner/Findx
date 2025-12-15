package utils

import "strings"

// TruncateString 截断字符串，如果超过 maxLength 则截断并添加省略号
func TruncateString(str string, maxLength int) string {
	// 移除字符串中的所有空格
	trimmedStr := strings.ReplaceAll(str, " ", "")

	// 如果移除空格后的字符串长度超过 maxLength，则截断
	if len(trimmedStr) > maxLength {
		return str[:maxLength] + "..."
	}
	return str
}
