package output

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"
)

//go:embed template/report.html
var templateFS embed.FS

// HTMLReport HTML报告数据结构
type HTMLReport struct {
	TotalFiles    int
	TotalFindings int
	Duration      string
	ScanTime      string
	GenerateTime  string
	ScanDirectory string
	CriticalCount int
	HighCount     int
	MediumCount   int
	LowCount      int
	Files         []HTMLFileSection
}

// HTMLFileSection 文件区域
type HTMLFileSection struct {
	Path    string
	Count   int
	Results []HTMLResult
}

// HTMLResult HTML结果项
type HTMLResult struct {
	Icon           string
	RuleName       string
	Type           string
	RiskLevel      string
	RiskLevelText  string
	MatchedValue   string
	LineNumber     string
	Offset         string
	Context        string
}

// HTMLReportGenerator HTML报告生成器
type HTMLReportGenerator struct {
	template *template.Template
}

// NewHTMLReportGenerator 创建HTML报告生成器
func NewHTMLReportGenerator() (*HTMLReportGenerator, error) {
	tmplContent, err := templateFS.ReadFile("template/report.html")
	if err != nil {
		return nil, fmt.Errorf("读取模板失败: %w", err)
	}

	tmpl, err := template.New("report").Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("解析模板失败: %w", err)
	}

	return &HTMLReportGenerator{
		template: tmpl,
	}, nil
}

// Generate 生成HTML报告
func (g *HTMLReportGenerator) Generate(outputPath string, report *HTMLReport) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建HTML文件失败: %w", err)
	}
	defer file.Close()

	// 写入 UTF-8 BOM
	file.Write([]byte{0xEF, 0xBB, 0xBF})

	if err := g.template.Execute(file, report); err != nil {
		return fmt.Errorf("生成HTML失败: %w", err)
	}

	return nil
}

// BuildHTMLReport 构建HTML报告数据
func BuildHTMLReport(scanDir string, duration time.Duration, fileResults map[string][]string) *HTMLReport {
	report := &HTMLReport{
		ScanDirectory: scanDir,
		Duration:      duration.String(),
		ScanTime:      time.Now().Format("2006-01-02 15:04:05"),
		GenerateTime:  time.Now().Format("2006-01-02 15:04:05"),
		Files:         make([]HTMLFileSection, 0),
	}

	// 处理每个文件的结果
	for filePath, results := range fileResults {
		if len(results) == 0 {
			continue
		}

		fileSection := HTMLFileSection{
			Path:    filePath,
			Count:   len(results),
			Results: make([]HTMLResult, 0),
		}

		for _, raw := range results {
			htmlResult := parseRawResult(raw)
			if htmlResult != nil {
				fileSection.Results = append(fileSection.Results, *htmlResult)
				
				// 统计风险等级
				switch strings.ToLower(htmlResult.RiskLevel) {
				case "critical":
					report.CriticalCount++
				case "high":
					report.HighCount++
				case "medium":
					report.MediumCount++
				case "low":
					report.LowCount++
				}
			}
		}

		report.Files = append(report.Files, fileSection)
		report.TotalFiles++
		report.TotalFindings += len(fileSection.Results)
	}

	return report
}

// parseRawResult 解析原始结果字符串
func parseRawResult(raw string) *HTMLResult {
	parts := strings.Split(raw, "|")
	if len(parts) < 2 {
		return nil
	}

	result := &HTMLResult{}

	switch parts[0] {
	case "TEXT":
		if len(parts) >= 4 {
			result.Icon = "🔑"
			result.RuleName = "关键字匹配: " + parts[1]
			result.Type = "文本文件"
			result.RiskLevel = "medium"
			result.RiskLevelText = "中危"
			result.LineNumber = parts[2]
			result.Context = parts[3]
			result.MatchedValue = parts[1]
		}

	case "WORD":
		if len(parts) >= 4 {
			result.Icon = "📄"
			result.RuleName = "关键字匹配: " + parts[2]
			result.Type = "Word文档 - " + parts[1]
			result.RiskLevel = "medium"
			result.RiskLevelText = "中危"
			result.Context = parts[3]
			result.MatchedValue = parts[2]
		}

	case "EXCEL":
		if len(parts) >= 4 {
			result.Icon = "📊"
			result.RuleName = "关键字匹配: " + parts[2]
			result.Type = "Excel文档 (" + parts[1] + ")"
			result.RiskLevel = "medium"
			result.RiskLevelText = "中危"
			result.Context = parts[3]
			result.MatchedValue = parts[2]
		}

	case "CSV":
		if len(parts) >= 3 {
			result.Icon = "📋"
			result.RuleName = "关键字匹配: " + parts[1]
			result.Type = "CSV文件"
			result.RiskLevel = "medium"
			result.RiskLevelText = "中危"
			result.Context = parts[2]
			result.MatchedValue = parts[1]
		}

	case "BINARY":
		if len(parts) >= 7 {
			result.Icon = getRiskIconText(parts[3])
			result.RuleName = parts[2]
			result.Type = parts[1]
			result.RiskLevel = strings.ToLower(parts[3])
			result.RiskLevelText = getRiskLevelText(parts[3])
			result.MatchedValue = parts[4]
			result.Offset = parts[5]
			result.Context = parts[6]
		}
	}

	return result
}

// getRiskIconText 获取风险图标文本
func getRiskIconText(riskLevel string) string {
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

// getRiskLevelText 获取风险等级文本
func getRiskLevelText(riskLevel string) string {
	switch strings.ToLower(riskLevel) {
	case "critical":
		return "严重"
	case "high":
		return "高危"
	case "medium":
		return "中危"
	case "low":
		return "低危"
	default:
		return "未知"
	}
}
