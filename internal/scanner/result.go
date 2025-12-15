package scanner

import (
	"fmt"
	"strings"
)

// ScanResult 统一的扫描结果结构
type ScanResult struct {
	Index        int    // 序号
	RuleName     string // 规则名称
	RiskLevel    string // 风险等级
	Confidence   string // 置信度
	Description  string // 规则描述
	MatchedValue string // 匹配值
	FilePath     string // 文件路径
	LineNumber   int    // 行号（文本文件）
	Offset       int    // 偏移量（二进制文件）
	Context      string // 上下文
	FileType     string // 文件类型（text/binary/document）
}

// FormatOutput 格式化输出结果
func (r *ScanResult) FormatOutput() string {
	var sb strings.Builder
	
	// 序号和规则名称
	sb.WriteString(fmt.Sprintf("%d. [%s] %s", r.Index, r.RuleName, r.RiskLevel))
	
	// 置信度（如果有）
	if r.Confidence != "" {
		sb.WriteString(fmt.Sprintf(" - %s", r.Confidence))
	}
	sb.WriteString("\n")
	
	// 规则描述
	if r.Description != "" {
		sb.WriteString(fmt.Sprintf("   规则描述: %s\n", r.Description))
	}
	
	// 匹配值
	sb.WriteString(fmt.Sprintf("   匹配值: %s\n", maskSensitiveValue(r.MatchedValue)))
	
	// 文件路径
	sb.WriteString(fmt.Sprintf("   文件: %s\n", r.FilePath))
	
	// 位置信息
	switch r.FileType {
	case "binary":
		sb.WriteString(fmt.Sprintf("   偏移: 0x%X\n", r.Offset))
	case "text", "document":
		if r.LineNumber > 0 {
			sb.WriteString(fmt.Sprintf("   行号: %d\n", r.LineNumber))
		}
	}
	
	// 上下文
	if r.Context != "" {
		sb.WriteString(fmt.Sprintf("   上下文: %s\n", r.Context))
	}
	
	return sb.String()
}

// FormatSimple 简单格式输出（用于实时输出）
func (r *ScanResult) FormatSimple() string {
	location := ""
	if r.FileType == "binary" {
		location = fmt.Sprintf("偏移:0x%X", r.Offset)
	} else if r.LineNumber > 0 {
		location = fmt.Sprintf("行:%d", r.LineNumber)
	}
	
	return fmt.Sprintf("[+] [%s] %s | %s | %s", 
		r.RuleName, 
		maskSensitiveValue(r.MatchedValue), 
		location,
		r.FilePath)
}

// maskSensitiveValue 对敏感值进行脱敏处理
func maskSensitiveValue(value string) string {
	if len(value) <= 6 {
		return strings.Repeat("*", len(value))
	}
	
	// 保留前2个和后2个字符
	prefix := value[:2]
	suffix := value[len(value)-2:]
	middle := strings.Repeat("*", len(value)-4)
	
	return prefix + middle + suffix
}

// GetRiskIcon 获取风险等级图标
func GetRiskIcon(riskLevel string) string {
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

// ResultCollection 结果集合
type ResultCollection struct {
	Results []ScanResult
	counter int
}

// NewResultCollection 创建结果集合
func NewResultCollection() *ResultCollection {
	return &ResultCollection{
		Results: make([]ScanResult, 0),
		counter: 0,
	}
}

// Add 添加结果
func (rc *ResultCollection) Add(result ScanResult) {
	rc.counter++
	result.Index = rc.counter
	rc.Results = append(rc.Results, result)
}

// Count 获取结果数量
func (rc *ResultCollection) Count() int {
	return len(rc.Results)
}

// GetStatistics 获取统计信息
func (rc *ResultCollection) GetStatistics() map[string]int {
	stats := make(map[string]int)
	stats["total"] = len(rc.Results)
	stats["critical"] = 0
	stats["high"] = 0
	stats["medium"] = 0
	stats["low"] = 0
	
	for _, result := range rc.Results {
		switch strings.ToLower(result.RiskLevel) {
		case "critical":
			stats["critical"]++
		case "high":
			stats["high"]++
		case "medium":
			stats["medium"]++
		case "low":
			stats["low"]++
		}
	}
	
	return stats
}

// PrintStatistics 打印统计信息
func (rc *ResultCollection) PrintStatistics() {
	stats := rc.GetStatistics()
	fmt.Printf("\n[*] 📊 扫描统计:\n")
	fmt.Printf("    总计: %d 个敏感信息\n", stats["total"])
	fmt.Printf("    🔴 严重: %d | 🟠 高危: %d | 🟡 中危: %d | 🟢 低危: %d\n",
		stats["critical"], stats["high"], stats["medium"], stats["low"])
}
