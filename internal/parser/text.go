package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// TextParser 文本文件解析器
type TextParser struct{}

// NewTextParser 创建文本解析器
func NewTextParser() *TextParser {
	return &TextParser{}
}

// Parse 解析文本文件内容
func (p *TextParser) Parse(filePath string, keywords []string, verbose bool) []string {
	var matchingLines []string
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("[-] 打开文件%s错误\n", filePath)
		return matchingLines
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				lineOutput := formatTextResult(keyword, lineNum, line)
				matchingLines = append(matchingLines, lineOutput)
				if verbose {
					fmt.Println(lineOutput)
				}
				break // 找到一个匹配的字段即可
			}
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[-] 读取文件错误%s: %v\n", filePath, err)
	}

	return matchingLines
}

// formatTextResult 格式化文本扫描结果
func formatTextResult(keyword string, lineNum int, content string) string {
	return fmt.Sprintf("TEXT|%s|%d|%s", keyword, lineNum, content)
}
