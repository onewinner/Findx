package parser

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// CSVParser CSV文件解析器
type CSVParser struct{}

// NewCSVParser 创建CSV解析器
func NewCSVParser() *CSVParser {
	return &CSVParser{}
}

// Parse 解析CSV文件内容
func (p *CSVParser) Parse(filePath string, keywords []string, verbose bool) []string {
	var matchingLines []string
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("[-] 打开CSV文件%s错误\n", filePath)
		return matchingLines
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("[-] 读取CSV文件%s错误\n", filePath)
		return matchingLines
	}

	for _, record := range records {
		for _, text := range record {
			for _, keyword := range keywords {
				if strings.Contains(text, keyword) {
					lineOutput := formatCSVResult(keyword, text)
					matchingLines = append(matchingLines, lineOutput)
					if verbose {
						fmt.Println(lineOutput)
					}
					break
				}
			}
		}
	}
	return matchingLines
}


// formatCSVResult 格式化CSV扫描结果
func formatCSVResult(keyword, content string) string {
	return fmt.Sprintf("CSV|%s|%s", keyword, content)
}
