package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

const Banner = `

    ___       ___       ___       ___       ___   
   /\  \     /\  \     /\__\     /\  \     /\__\  
  /::\  \   _\:\  \   /:| _|_   /::\  \   |::L__L 
 /::\:\__\ /\/::\__\ /::|/\__\ /:/\:\__\ /::::\__\
 \/\:\/__/ \::/\/__/ \/|::/  / \:\/:/  / \;::;/__/
    \/__/   \:\__\     |:/  /   \::/  /   |::|__| 
             \/__/     \/__/     \/__/     \/__/  
   
`

// Config 扫描配置
type Config struct {
	// 基础配置
	FileTypes   []string // 文件类型列表
	Keywords    []string // 搜索关键词列表
	OutputFile  string   // 输出文件路径
	HTMLOutput  string   // HTML报告文件路径
	Directory   string   // 扫描目录
	Verbose     bool     // 是否实时输出
	ThreadCount int      // 线程数
	
	// 高级配置
	MaxFileSize  int64    // 最大文件大小（字节）
	ExcludeDirs  []string // 排除目录列表
	ExcludeFiles []string // 排除文件模式列表
	
	// 二进制扫描配置
	BinaryMode    bool // 是否启用二进制扫描模式
	ContextLength int  // 上下文长度
}

// Validate 验证配置有效性
func (c *Config) Validate() error {
	if c.Directory == "" {
		return fmt.Errorf("扫描目录不能为空")
	}
	
	if len(c.FileTypes) == 0 {
		return fmt.Errorf("文件类型列表不能为空")
	}
	
	// 二进制模式下，关键词可以为空（只使用规则匹配）
	// 文本模式下，关键词不能为空
	if len(c.Keywords) == 0 && !c.BinaryMode && !c.HasBinaryFileTypes() {
		return fmt.Errorf("关键词列表不能为空（除非启用二进制扫描模式）")
	}
	
	if c.ThreadCount < 1 {
		return fmt.Errorf("线程数必须大于0")
	}
	
	return nil
}

// ShouldExcludeDir 判断是否应该排除该目录
func (c *Config) ShouldExcludeDir(dirPath string) bool {
	if len(c.ExcludeDirs) == 0 {
		return false
	}
	
	dirName := filepath.Base(dirPath)
	for _, exclude := range c.ExcludeDirs {
		if dirName == exclude || strings.Contains(dirPath, exclude) {
			return true
		}
	}
	
	return false
}

// ShouldExcludeFile 判断是否应该排除该文件
func (c *Config) ShouldExcludeFile(filePath string) bool {
	if len(c.ExcludeFiles) == 0 {
		return false
	}
	
	fileName := filepath.Base(filePath)
	for _, pattern := range c.ExcludeFiles {
		// 简单的通配符匹配
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
	}
	
	return false
}

// ShouldSkipBySize 判断文件是否因大小超限而跳过
func (c *Config) ShouldSkipBySize(fileSize int64) bool {
	if c.MaxFileSize <= 0 {
		return false
	}
	return fileSize > c.MaxFileSize
}

// PrintConfig 打印配置信息
func (c *Config) PrintConfig() {
	fmt.Println("[*] 扫描配置:")
	fmt.Printf("    目录: %s\n", c.Directory)
	fmt.Printf("    输出: %s\n", c.OutputFile)
	fmt.Printf("    线程: %d\n", c.ThreadCount)
	fmt.Printf("    文件类型: %s\n", strings.Join(c.FileTypes, ", "))
	
	// 显示关键词信息
	if len(c.Keywords) > 0 {
		fmt.Printf("    关键词数: %d 个\n", len(c.Keywords))
	} else {
		fmt.Println("    关键词: 无（仅使用规则匹配）")
	}
	
	// 自动检测二进制文件类型
	if c.BinaryMode || c.HasBinaryFileTypes() {
		binaryTypes := c.GetBinaryFileTypes()
		if len(binaryTypes) > 0 {
			fmt.Printf("    模式: 二进制扫描模式 (%s)\n", strings.Join(binaryTypes, ", "))
		} else {
			fmt.Println("    模式: 二进制扫描模式 (DLL/EXE/SO)")
		}
	}
	
	if c.MaxFileSize > 0 {
		fmt.Printf("    最大文件: %.2f MB\n", float64(c.MaxFileSize)/1024/1024)
	}
	
	if len(c.ExcludeDirs) > 0 {
		fmt.Printf("    排除目录: %s\n", strings.Join(c.ExcludeDirs, ", "))
	}
	
	if len(c.ExcludeFiles) > 0 {
		fmt.Printf("    排除文件: %s\n", strings.Join(c.ExcludeFiles, ", "))
	}
	
	fmt.Println("[*] 🚀🚀🚀🚀🚀🚀开始扫描🚀🚀🚀🚀🚀🚀")
}

// GetFileTypeCount 获取文件类型数量
func (c *Config) GetFileTypeCount() int {
	return len(c.FileTypes)
}

// GetKeywordCount 获取关键词数量
func (c *Config) GetKeywordCount() int {
	return len(c.Keywords)
}

// IsFileTypeSupported 判断文件类型是否支持
func (c *Config) IsFileTypeSupported(filePath string) bool {
	for _, ext := range c.FileTypes {
		if strings.HasSuffix(filePath, ext) {
			return true
		}
	}
	return false
}

// HasBinaryFileTypes 检查配置中是否包含二进制文件类型
func (c *Config) HasBinaryFileTypes() bool {
	binaryExts := []string{".dll", ".exe", ".so", ".dylib", ".bin", ".o", ".obj"}
	
	for _, fileType := range c.FileTypes {
		fileTypeLower := strings.ToLower(fileType)
		for _, binaryExt := range binaryExts {
			if fileTypeLower == binaryExt {
				return true
			}
		}
	}
	return false
}

// GetBinaryFileTypes 获取配置中的二进制文件类型
func (c *Config) GetBinaryFileTypes() []string {
	binaryExts := []string{".dll", ".exe", ".so", ".dylib", ".bin", ".o", ".obj"}
	var result []string
	
	for _, fileType := range c.FileTypes {
		fileTypeLower := strings.ToLower(fileType)
		for _, binaryExt := range binaryExts {
			if fileTypeLower == binaryExt {
				result = append(result, fileType)
				break
			}
		}
	}
	return result
}
