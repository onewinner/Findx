package parser

import (
	"fmt"
	"strings"

	"github.com/carmel/gooxml/document"
)

// WordParser Word文档解析器
type WordParser struct{}

// NewWordParser 创建Word解析器
func NewWordParser() *WordParser {
	return &WordParser{}
}

// Parse 解析Word文档内容
func (p *WordParser) Parse(filePath string, keywords []string, verbose bool) []string {
	var matchingLines []string
	doc, err := document.Open(filePath)
	if err != nil {
		fmt.Printf("[-] 打开Word文件%s错误\n", filePath)
		return matchingLines
	}

	// 搜索段落
	for _, para := range doc.Paragraphs() {
		for _, run := range para.Runs() {
			text := run.Text()
			for _, keyword := range keywords {
				if strings.Contains(text, keyword) {
					lineOutput := formatWordResult(keyword, "段落", text)
					matchingLines = append(matchingLines, lineOutput)
					if verbose {
						fmt.Println(lineOutput)
					}
					break
				}
			}
		}
	}

	// 搜索表格
	for _, table := range doc.Tables() {
		for _, row := range table.Rows() {
			for _, cell := range row.Cells() {
				for _, para := range cell.Paragraphs() {
					for _, run := range para.Runs() {
						text := run.Text()
						for _, keyword := range keywords {
							if strings.Contains(text, keyword) {
								lineOutput := formatWordResult(keyword, "表格", text)
								matchingLines = append(matchingLines, lineOutput)
								if verbose {
									fmt.Println(lineOutput)
								}
								break
							}
						}
					}
				}
			}
		}
	}

	return matchingLines
}

// formatWordResult 格式化Word扫描结果
func formatWordResult(keyword, location, content string) string {
	return fmt.Sprintf("WORD|%s|%s|%s", location, keyword, content)
}
