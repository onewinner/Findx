package output

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Writer 输出写入器
type Writer struct {
	outputFile string
	file       *os.File
	writer     *bufio.Writer
}

// NewWriter 创建输出写入器
func NewWriter(outputFile string) *Writer {
	return &Writer{
		outputFile: outputFile,
	}
}

// Open 打开输出文件
func (w *Writer) Open() error {
	file, err := os.OpenFile(w.outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开输出文件失败: %w", err)
	}
	
	// 如果是新文件，写入 UTF-8 BOM
	fileInfo, _ := file.Stat()
	if fileInfo.Size() == 0 {
		// UTF-8 BOM: EF BB BF
		file.Write([]byte{0xEF, 0xBB, 0xBF})
	}
	
	w.file = file
	w.writer = bufio.NewWriter(file)
	return nil
}

// Close 关闭输出文件
func (w *Writer) Close() error {
	if w.writer != nil {
		w.writer.Flush()
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// WriteResults 将匹配结果写入文件（旧格式，保持兼容）
func (w *Writer) WriteResults(filePath string, matchingLines []string) error {
	file, err := os.OpenFile(w.outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开输出文件失败: %w", err)
	}
	defer file.Close()

	// 如果是新文件，写入 UTF-8 BOM
	fileInfo, _ := file.Stat()
	if fileInfo.Size() == 0 {
		// UTF-8 BOM: EF BB BF
		file.Write([]byte{0xEF, 0xBB, 0xBF})
	}

	writer := bufio.NewWriter(file)
	fmt.Fprintf(writer, "[!] 文件地址: %s\n", filePath)
	for _, line := range matchingLines {
		fmt.Fprintln(writer, line)
	}
	fmt.Fprintln(writer)
	
	return writer.Flush()
}

// WriteResult 写入单个结果（新格式）
func (w *Writer) WriteResult(result string) error {
	if w.writer == nil {
		return fmt.Errorf("输出文件未打开")
	}
	
	_, err := w.writer.WriteString(result)
	if err != nil {
		return err
	}
	
	_, err = w.writer.WriteString(strings.Repeat("-", 80) + "\n\n")
	return err
}

// Flush 刷新缓冲区
func (w *Writer) Flush() error {
	if w.writer != nil {
		return w.writer.Flush()
	}
	return nil
}

// WriteFormattedResults 写入格式化的结果
func (w *Writer) WriteFormattedResults(results []string) error {
	file, err := os.OpenFile(w.outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开输出文件失败: %w", err)
	}
	defer file.Close()

	// 如果是新文件，写入 UTF-8 BOM
	fileInfo, _ := file.Stat()
	if fileInfo.Size() == 0 {
		// UTF-8 BOM: EF BB BF
		file.Write([]byte{0xEF, 0xBB, 0xBF})
	}

	writer := bufio.NewWriter(file)
	for _, result := range results {
		fmt.Fprint(writer, result)
	}

	return writer.Flush()
}
