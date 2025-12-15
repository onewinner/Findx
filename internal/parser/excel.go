package parser

import (
	"fmt"
	"strings"

	"github.com/extrame/xls"
	"github.com/tealeg/xlsx"
)

// ExcelParser Excel文档解析器
type ExcelParser struct{}

// NewExcelParser 创建Excel解析器
func NewExcelParser() *ExcelParser {
	return &ExcelParser{}
}

// ParseXLSX 解析.xlsx文件
func (p *ExcelParser) ParseXLSX(filePath string, keywords []string, verbose bool) []string {
	var matchingLines []string
	xlFile, err := xlsx.OpenFile(filePath)
	if err != nil {
		fmt.Printf("[-] 打开Excel文件%s错误\n", filePath)
		return matchingLines
	}

	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				text := cell.String()
				for _, keyword := range keywords {
					if strings.Contains(text, keyword) {
						lineOutput := formatExcelResult(keyword, "XLSX", text)
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
	return matchingLines
}

// ParseXLS 解析.xls文件
func (p *ExcelParser) ParseXLS(filePath string, keywords []string, verbose bool) []string {
	var matchingLines []string
	xlFile, err := xls.Open(filePath, "utf-8")
	if err != nil {
		fmt.Printf("[-] 打开XLS文件%s错误\n", filePath)
		return matchingLines
	}

	for i := 0; i < xlFile.NumSheets(); i++ {
		sheet := xlFile.GetSheet(i)
		for j := 0; j <= int(sheet.MaxRow); j++ {
			row := sheet.Row(j)
			for k := 0; k < row.LastCol(); k++ {
				text := row.Col(k)
				for _, keyword := range keywords {
					if strings.Contains(text, keyword) {
						lineOutput := formatExcelResult(keyword, "XLS", text)
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
	return matchingLines
}


// formatExcelResult 格式化Excel扫描结果
func formatExcelResult(keyword, fileType, content string) string {
	return fmt.Sprintf("EXCEL|%s|%s|%s", fileType, keyword, content)
}
