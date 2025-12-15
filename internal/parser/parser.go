package parser

import (
	"os"
	"strings"
)

// Parser 文件解析器接口
type Parser interface {
	Parse(filePath string, keywords []string, verbose bool) []string
}

// ParserConfig 解析器配置
type ParserConfig struct {
	ContextLength int
}

// FileParser 文件解析器管理器
type FileParser struct {
	textParser    *TextParser
	wordParser    *WordParser
	excelParser   *ExcelParser
	csvParser     *CSVParser
	binaryParser  *BinaryParser
	contextLength int
}

// NewFileParser 创建文件解析器管理器
func NewFileParser(contextLength int) *FileParser {
	return &FileParser{
		textParser:    NewTextParser(),
		wordParser:    NewWordParser(),
		excelParser:   NewExcelParser(),
		csvParser:     NewCSVParser(),
		binaryParser:  NewBinaryParser(),
		contextLength: contextLength,
	}
}

// Parse 根据文件类型选择合适的解析器
func (fp *FileParser) Parse(filePath string, keywords []string, verbose bool) []string {
	// 检查是否为二进制文件（DLL/EXE）
	if isBinaryFile(filePath) {
		return fp.parseBinaryFile(filePath, keywords, verbose)
	}

	// 文档文件
	switch {
	case strings.HasSuffix(filePath, ".docx"):
		return fp.wordParser.Parse(filePath, keywords, verbose)
	case strings.HasSuffix(filePath, ".xlsx"):
		return fp.excelParser.ParseXLSX(filePath, keywords, verbose)
	case strings.HasSuffix(filePath, ".xls"):
		return fp.excelParser.ParseXLS(filePath, keywords, verbose)
	case strings.HasSuffix(filePath, ".csv"):
		return fp.csvParser.Parse(filePath, keywords, verbose)
	default:
		return fp.textParser.Parse(filePath, keywords, verbose)
	}
}

// isBinaryFile 判断是否为二进制文件
func isBinaryFile(filePath string) bool {
	ext := strings.ToLower(filePath)
	// 支持的二进制文件扩展名
	binaryExts := []string{".dll", ".exe", ".so", ".dylib", ".bin", ".o", ".obj"}
	
	for _, binaryExt := range binaryExts {
		if strings.HasSuffix(ext, binaryExt) {
			return true
		}
	}
	return false
}

// parseBinaryFile 解析二进制文件
func (fp *FileParser) parseBinaryFile(filePath string, keywords []string, verbose bool) []string {
	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		if verbose {
			println("[-] 读取二进制文件失败:", filePath)
		}
		return nil
	}

	// 使用二进制解析器（带关键字和上下文长度）
	return fp.binaryParser.ParseWithKeywords(filePath, data, keywords, verbose, fp.contextLength)
}
