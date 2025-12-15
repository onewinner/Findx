package output

import (
	"fmt"
	"strings"
)

// ResultFormatter 结果格式化器
type ResultFormatter struct {
	width int // 输出宽度
}

// NewResultFormatter 创建格式化器
func NewResultFormatter() *ResultFormatter {
	return &ResultFormatter{
		width: 100, // 默认宽度
	}
}

// FormatFileHeader 格式化文件头
func (f *ResultFormatter) FormatFileHeader(filePath string, count int) string {
	var sb strings.Builder
	
	sb.WriteString("\n")
	sb.WriteString(f.line("═"))
	sb.WriteString(f.centerLine(fmt.Sprintf("📄 文件: %s", truncatePath(filePath, 80))))
	sb.WriteString(f.centerLine(fmt.Sprintf("🔍 发现 %d 个敏感信息", count)))
	sb.WriteString(f.line("═"))
	sb.WriteString("\n")
	
	return sb.String()
}

// FormatBinaryResult 格式化二进制扫描结果
func (f *ResultFormatter) FormatBinaryResult(index int, matchType, ruleName, riskLevel, matchedValue string, offset int, context string) string {
	var sb strings.Builder
	
	riskIcon := getRiskIcon(riskLevel)
	
	sb.WriteString(fmt.Sprintf("\n[%d] %s %s\n", index, riskIcon, ruleName))
	sb.WriteString(f.line("─"))
	sb.WriteString(fmt.Sprintf("  类型: %s\n", matchType))
	sb.WriteString(fmt.Sprintf("  风险: %s %s\n", riskIcon, riskLevel))
	sb.WriteString(fmt.Sprintf("  匹配: %s\n", matchedValue))
	
	// 只有当偏移有效时才显示
	if offset >= 0 {
		sb.WriteString(fmt.Sprintf("  偏移: 0x%X\n", offset))
	}
	
	sb.WriteString(fmt.Sprintf("  上下文:\n"))
	sb.WriteString(f.wrapText(context, "    "))
	sb.WriteString("\n")
	
	return sb.String()
}

// FormatTextResult 格式化文本扫描结果
func (f *ResultFormatter) FormatTextResult(index int, keyword string, lineNum int, content string) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("\n[%d] 🔑 关键字匹配: %s\n", index, keyword))
	sb.WriteString(f.line("─"))
	sb.WriteString(fmt.Sprintf("  类型: 文本文件\n"))
	sb.WriteString(fmt.Sprintf("  行号: %d\n", lineNum))
	sb.WriteString(fmt.Sprintf("  内容:\n"))
	sb.WriteString(f.wrapText(content, "    "))
	sb.WriteString("\n")
	
	return sb.String()
}

// FormatDocumentResult 格式化文档扫描结果
func (f *ResultFormatter) FormatDocumentResult(index int, docType, location, keyword, content string) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("\n[%d] 📋 关键字匹配: %s\n", index, keyword))
	sb.WriteString(f.line("─"))
	sb.WriteString(fmt.Sprintf("  类型: %s\n", docType))
	sb.WriteString(fmt.Sprintf("  位置: %s\n", location))
	sb.WriteString(fmt.Sprintf("  内容:\n"))
	sb.WriteString(f.wrapText(content, "    "))
	sb.WriteString("\n")
	
	return sb.String()
}

// FormatSummary 格式化扫描摘要
func (f *ResultFormatter) FormatSummary(totalFiles, totalFindings int, elapsed string, stats map[string]int) string {
	var sb strings.Builder
	
	sb.WriteString("\n")
	sb.WriteString(f.line("═"))
	sb.WriteString(f.centerLine("📊 扫描完成"))
	sb.WriteString(f.line("═"))
	sb.WriteString(fmt.Sprintf("  扫描文件: %d 个\n", totalFiles))
	sb.WriteString(fmt.Sprintf("  发现问题: %d 个\n", totalFindings))
	sb.WriteString(fmt.Sprintf("  耗时: %s\n", elapsed))
	
	if len(stats) > 0 {
		sb.WriteString(fmt.Sprintf("\n  风险分布:\n"))
		if stats["critical"] > 0 {
			sb.WriteString(fmt.Sprintf("    🔴 严重: %d\n", stats["critical"]))
		}
		if stats["high"] > 0 {
			sb.WriteString(fmt.Sprintf("    🟠 高危: %d\n", stats["high"]))
		}
		if stats["medium"] > 0 {
			sb.WriteString(fmt.Sprintf("    🟡 中危: %d\n", stats["medium"]))
		}
		if stats["low"] > 0 {
			sb.WriteString(fmt.Sprintf("    🟢 低危: %d\n", stats["low"]))
		}
	}
	
	sb.WriteString(f.line("═"))
	sb.WriteString("\n")
	
	return sb.String()
}

// line 生成分隔线
func (f *ResultFormatter) line(char string) string {
	return strings.Repeat(char, f.width) + "\n"
}

// centerLine 生成居中文本行
func (f *ResultFormatter) centerLine(text string) string {
	// 计算实际字符宽度（中文字符算2个宽度）
	textWidth := displayWidth(text)
	if textWidth >= f.width {
		return text + "\n"
	}
	
	padding := (f.width - textWidth) / 2
	return strings.Repeat(" ", padding) + text + "\n"
}

// wrapText 文本换行
func (f *ResultFormatter) wrapText(text, prefix string) string {
	maxWidth := f.width - len(prefix) - 2
	if len(text) <= maxWidth {
		return prefix + text + "\n"
	}
	
	var sb strings.Builder
	lines := splitByWidth(text, maxWidth)
	for _, line := range lines {
		sb.WriteString(prefix + line + "\n")
	}
	return sb.String()
}

// getRiskIcon 获取风险图标
func getRiskIcon(riskLevel string) string {
	switch strings.ToLower(riskLevel) {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	default:
		return "⚪"
	}
}

// truncatePath 截断路径
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}

// displayWidth 计算显示宽度（中文字符算2个宽度）
func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		if r > 127 {
			width += 2 // 中文字符
		} else {
			width += 1 // ASCII字符
		}
	}
	return width
}

// splitByWidth 按宽度分割文本
func splitByWidth(text string, maxWidth int) []string {
	var lines []string
	var currentLine strings.Builder
	currentWidth := 0
	
	for _, r := range text {
		charWidth := 1
		if r > 127 {
			charWidth = 2
		}
		
		if currentWidth+charWidth > maxWidth && currentLine.Len() > 0 {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentWidth = 0
		}
		
		currentLine.WriteRune(r)
		currentWidth += charWidth
	}
	
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	
	return lines
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
