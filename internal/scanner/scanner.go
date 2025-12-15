package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"Findx/internal/config"
	"Findx/internal/output"
	"Findx/internal/parser"
)

// Scanner 文件扫描器
type Scanner struct {
	config      *config.Config
	fileParser  *parser.FileParser
	writer      *output.Writer
	fileResults map[string][]string // 收集每个文件的结果用于生成HTML
	mu          sync.Mutex          // 保护 fileResults
}

// NewScanner 创建扫描器
func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{
		config:      cfg,
		fileParser:  parser.NewFileParser(cfg.ContextLength),
		writer:      output.NewWriter(cfg.OutputFile),
		fileResults: make(map[string][]string),
	}
}

// Run 执行扫描
func (s *Scanner) Run() error {
	start := time.Now()

	// 搜索文件
	files := s.searchFiles()
	if len(files) == 0 {
		fmt.Println("[*] 未找到匹配的文件")
		return nil
	}

	// 使用工作池进行并发扫描
	s.scanFiles(files)

	// 输出统计信息
	elapsed := time.Since(start)
	fmt.Printf("[*] 🎉🎉🎉🎉🎉🎉扫描完成🎉🎉🎉🎉🎉🎉\n")
	fmt.Printf("[*] 扫描文件总数: %d    总耗时: %s\n", len(files), elapsed)
	fmt.Printf("[*] 详细结果保存至: %s\n", s.config.OutputFile)
	
	// 生成HTML报告
	if err := s.generateHTMLReport(elapsed); err != nil {
		fmt.Printf("[-] 生成HTML报告失败: %v\n", err)
	} else {
		fmt.Printf("[*] HTML报告保存至: %s\n", s.config.HTMLOutput)
	}

	return nil
}

// searchFiles 搜索目录中的文件
func (s *Scanner) searchFiles() []string {
	var files []string
	var skippedDirs int
	var skippedFiles int
	var skippedSize int
	
	err := filepath.Walk(s.config.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 检查是否排除目录
		if info.IsDir() {
			if s.config.ShouldExcludeDir(path) {
				skippedDirs++
				if s.config.Verbose {
					fmt.Printf("[*] 跳过目录: %s\n", path)
				}
				return filepath.SkipDir
			}
			return nil
		}
		
		// 检查是否排除文件
		if s.config.ShouldExcludeFile(path) {
			skippedFiles++
			return nil
		}
		
		// 检查文件大小
		if s.config.ShouldSkipBySize(info.Size()) {
			skippedSize++
			if s.config.Verbose {
				fmt.Printf("[*] 跳过大文件: %s (%.2f MB)\n", path, float64(info.Size())/1024/1024)
			}
			return nil
		}
		
		// 检查文件类型
		if s.config.IsFileTypeSupported(info.Name()) {
			files = append(files, path)
		}
		
		return nil
	})
	
	if err != nil {
		fmt.Printf("[-] 扫描目录错误: %v\n", err)
	}
	
	// 打印统计信息
	if skippedDirs > 0 || skippedFiles > 0 || skippedSize > 0 {
		fmt.Printf("[*] 跳过统计: 目录(%d) 文件(%d) 大文件(%d)\n", skippedDirs, skippedFiles, skippedSize)
	}
	
	return files
}

// scanFiles 并发扫描文件
func (s *Scanner) scanFiles(files []string) {
	var wg sync.WaitGroup
	var mu sync.Mutex // 添加互斥锁保护输出
	semaphore := make(chan struct{}, s.config.ThreadCount)
	
	formatter := output.NewResultFormatter()
	resultIndex := 0

	for _, filePath := range files {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			
			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 解析文件内容
			rawResults := s.fileParser.Parse(path, s.config.Keywords, false) // 关闭原始输出
			
			// 写入结果
			if len(rawResults) > 0 {
				// 使用互斥锁保护输出，确保同一文件的结果不被打断
				mu.Lock()
				defer mu.Unlock()
				
				// 收集结果用于HTML报告
				s.mu.Lock()
				s.fileResults[path] = rawResults
				s.mu.Unlock()
				
				// 格式化文件头
				header := formatter.FormatFileHeader(path, len(rawResults))
				
				// 如果启用了 verbose，先输出文件头到控制台
				if s.config.Verbose {
					fmt.Print(header)
				}
				
				// 格式化每个结果
				var formattedResults []string
				formattedResults = append(formattedResults, header)
				
				for _, raw := range rawResults {
					resultIndex++
					formatted := s.formatResult(formatter, resultIndex, raw)
					formattedResults = append(formattedResults, formatted)
					
					// 如果启用了 verbose，输出格式化后的结果到控制台
					if s.config.Verbose {
						fmt.Print(formatted)
					}
				}
				
				if err := s.writer.WriteFormattedResults(formattedResults); err != nil {
					fmt.Printf("[-] 写入结果失败: %v\n", err)
				}
			}
		}(filePath)
	}

	wg.Wait()
}

// formatResult 格式化单个结果
func (s *Scanner) formatResult(formatter *output.ResultFormatter, index int, raw string) string {
	parts := strings.Split(raw, "|")
	if len(parts) < 2 {
		return raw
	}
	
	switch parts[0] {
	case "TEXT":
		if len(parts) >= 4 {
			keyword := parts[1]
			lineNum := 0
			fmt.Sscanf(parts[2], "%d", &lineNum)
			content := parts[3]
			return formatter.FormatTextResult(index, keyword, lineNum, content)
		}
	case "WORD":
		if len(parts) >= 4 {
			location := parts[1]
			keyword := parts[2]
			content := parts[3]
			return formatter.FormatDocumentResult(index, "Word文档", location, keyword, content)
		}
	case "EXCEL":
		if len(parts) >= 4 {
			fileType := parts[1]
			keyword := parts[2]
			content := parts[3]
			return formatter.FormatDocumentResult(index, fmt.Sprintf("Excel文档 (%s)", fileType), "单元格", keyword, content)
		}
	case "CSV":
		if len(parts) >= 3 {
			keyword := parts[1]
			content := parts[2]
			return formatter.FormatDocumentResult(index, "CSV文件", "字段", keyword, content)
		}
	case "BINARY":
		if len(parts) >= 7 {
			matchType := parts[1]
			ruleName := parts[2]
			riskLevel := parts[3]
			matchedValue := parts[4]
			offset := 0
			fmt.Sscanf(parts[5], "0x%X", &offset)
			context := parts[6]
			return formatter.FormatBinaryResult(index, matchType, ruleName, riskLevel, matchedValue, offset, context)
		}
	}
	
	return raw
}


// truncateForBox 截断字符串以适应框格
func truncateForBox(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return "..." + s[len(s)-maxLen+3:]
}

// generateHTMLReport 生成HTML报告
func (s *Scanner) generateHTMLReport(duration time.Duration) error {
	// 创建HTML报告生成器
	generator, err := output.NewHTMLReportGenerator()
	if err != nil {
		return err
	}
	
	// 构建报告数据
	report := output.BuildHTMLReport(s.config.Directory, duration, s.fileResults)
	
	// 使用配置中的HTML输出路径
	return generator.Generate(s.config.HTMLOutput, report)
}
