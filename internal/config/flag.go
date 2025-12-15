package config

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"
)

// 默认配置常量
const (
	DefaultFileTypes = ".txt,.log,.ini,.conf,.yaml,.yml,.xml,.json,.sql,.properties,.md,.java,.docx, .xlsx, .xls, .csv"
	DefaultKeywords  = "password=,username=,jdbc:,user=,ssh-,ldap:,mysqli_connect,sk-,账号,密码,username:,password:"
	DefaultOutput    = "res.txt"

	// 二进制文件类型
	BinaryFileTypes = ".dll,.exe,.so,.dylib,.bin,.o,.obj"
)

// GetFlags 返回所有命令行标志
func GetFlags() []cli.Flag {
	return []cli.Flag{
		// 基础参数
		&cli.StringFlag{
			Name:     "f",
			Aliases:  []string{"folder"},
			Usage:    "扫描目录（必填） / Scan directory (required)",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "o",
			Aliases: []string{"output"},
			Usage:   "输出文件路径 / Output file path",
			Value:   DefaultOutput,
		},
		&cli.StringFlag{
			Name:    "html",
			Aliases: []string{"html-output"},
			Usage:   "HTML报告文件路径（默认为输出文件名.html） / HTML report file path (default: output_file.html)",
		},

		// 文件类型参数
		&cli.StringFlag{
			Name:    "t",
			Aliases: []string{"type"},
			Usage:   "指定文件类型（逗号分隔） / Specify file types (comma separated)",
			Value:   DefaultFileTypes,
		},
		&cli.StringFlag{
			Name:    "ta",
			Aliases: []string{"type-append"},
			Usage:   "追加文件类型（逗号分隔） / Append file types (comma separated)",
		},

		// 关键词参数
		&cli.StringFlag{
			Name:    "k",
			Aliases: []string{"keyword"},
			Usage:   "搜索关键词（逗号分隔，二进制模式下可为空） / Search keywords (comma separated, can be empty in binary mode)",
			Value:   DefaultKeywords,
		},
		&cli.StringFlag{
			Name:    "ka",
			Aliases: []string{"keyword-append"},
			Usage:   "追加关键词（逗号分隔） / Append keywords (comma separated)",
		},

		// 性能参数
		&cli.IntFlag{
			Name:    "n",
			Aliases: []string{"thread"},
			Usage:   "线程数 / Number of threads",
			Value:   runtime.NumCPU(),
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"vb"},
			Usage:   "实时输出扫描结果 / Real-time output scan results",
			Value:   true,
		},

		// 高级参数
		&cli.Int64Flag{
			Name:    "s",
			Aliases: []string{"max-size"},
			Usage:   "最大文件大小（MB，0表示不限制） / Max file size (MB, 0 means no limit)",
			Value:   0,
		},
		&cli.StringFlag{
			Name:    "ed",
			Aliases: []string{"exclude-dir"},
			Usage:   "排除目录（逗号分隔） / Exclude directories (comma separated)",
		},
		&cli.StringFlag{
			Name:    "ef",
			Aliases: []string{"exclude-file"},
			Usage:   "排除文件模式（逗号分隔） / Exclude file patterns (comma separated)",
		},

		// 二进制扫描参数
		&cli.BoolFlag{
			Name:    "b",
			Aliases: []string{"binary"},
			Usage:   "启用二进制文件扫描模式（DLL/EXE） / Enable binary file scan mode (DLL/EXE)",
			Value:   false,
		},
		&cli.IntFlag{
			Name:    "ctx",
			Aliases: []string{"context"},
			Usage:   "上下文长度（字符数） / Context length (characters)",
			Value:   150,
		},
	}
}

// ParseConfig 从 cli.Context 解析配置
func ParseConfig(c *cli.Context) (*Config, error) {
	// 获取基础参数
	directory := c.String("f")
	output := c.String("o")

	// 合并文件类型
	fileTypes := parseList(c.String("t"))
	if appendTypes := c.String("ta"); appendTypes != "" {
		fileTypes = append(fileTypes, parseList(appendTypes)...)
	}

	// 如果启用二进制模式，添加二进制文件类型
	if c.Bool("b") {
		binaryTypes := parseList(BinaryFileTypes)
		fileTypes = append(fileTypes, binaryTypes...)
	}

	// 合并关键词
	keywords := parseList(c.String("k"))
	if appendKeywords := c.String("ka"); appendKeywords != "" {
		keywords = append(keywords, parseList(appendKeywords)...)
	}

	// 解析排除规则
	excludeDirs := parseList(c.String("ed"))
	excludeFiles := parseList(c.String("ef"))

	// 获取性能参数
	threadCount := c.Int("n")
	if threadCount < 1 {
		threadCount = 1
	}

	// 获取HTML输出路径
	htmlOutput := c.String("html")
	if htmlOutput == "" {
		// 如果没有指定，默认为输出文件名.html
		htmlOutput = strings.TrimSuffix(output, ".txt") + ".html"
	}

	// 创建配置对象
	config := &Config{
		FileTypes:     fileTypes,
		Keywords:      keywords,
		OutputFile:    output,
		HTMLOutput:    htmlOutput,
		Directory:     directory,
		Verbose:       c.Bool("verbose"),
		ThreadCount:   threadCount,
		MaxFileSize:   c.Int64("s") * 1024 * 1024, // 转换为字节
		ExcludeDirs:   excludeDirs,
		ExcludeFiles:  excludeFiles,
		BinaryMode:    c.Bool("b"),
		ContextLength: c.Int("ctx"),
	}

	return config, nil
}

// parseList 解析逗号分隔的列表
func parseList(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

// GetAppInfo 返回应用信息
func GetAppInfo() (name, usage, version string) {
	return "Findx",
		"文件敏感信息扫描工具 / File sensitive information scanner",
		"v1.0.0"
}

// GetDescription 返回详细描述
func GetDescription() string {
	return `Findx 是一个强大的文件敏感信息扫描工具，支持：
  - 多种文件格式：文本、Word、Excel、CSV、二进制文件（DLL/EXE）
  - 并发扫描：充分利用多核CPU
  - 灵活配置：支持自定义文件类型和关键词
  - 高级过滤：支持排除目录和文件、限制文件大小

Findx is a powerful file sensitive information scanner that supports:
  - Multiple file formats: text, Word, Excel, CSV, binary files (DLL/EXE)
  - Concurrent scanning: fully utilize multi-core CPU
  - Flexible configuration: support custom file types and keywords
  - Advanced filtering: support exclude directories and files, limit file size`
}

// GetExamples 返回使用示例
func GetExamples() string {
	return `示例 / Examples:

  # 基本扫描 / Basic scan
  findx -f /path/to/scan

  # 指定文件类型和关键词 / Specify file types and keywords
  findx -f /path/to/scan -t .txt,.log -k "password,token"

  # 自定义输出文件和HTML报告名称 / Custom output and HTML report names
  findx -f /path/to/scan -o result.txt --html report.html

  # 扫描Java项目 / Scan Java project
  findx -f /path/to/java-project -t .java,.properties,.xml -k "password,jdbc"

  # 扫描Python项目，排除虚拟环境 / Scan Python project, exclude virtual environment
  findx -f /path/to/python-project -t .py,.ini,.yaml -ed "venv,.venv"

  # 扫描二进制文件（方式1：使用 -b 参数）/ Scan binary files (method 1: use -b flag)
  findx -b -f /path/to/binaries

  # 扫描二进制文件（方式2：直接指定文件类型）/ Scan binary files (method 2: specify file types)
  findx -t .dll,.exe -f /path/to/binaries

  # 扫描二进制文件（方式3：追加二进制类型）/ Scan binary files (method 3: append binary types)
  findx -ta .dll,.exe -f /path/to/binaries

  # 扫描二进制文件，只使用规则匹配（不使用关键字）/ Scan binary files with rules only (no keywords)
  findx -b -k "" -f /path/to/binaries

  # 扫描二进制文件并自定义上下文长度 / Scan binary files with custom context length
  findx -b -f /path/to/binaries --ctx 200

  # 高性能扫描 / High performance scan
  findx -f /path/to/scan -n 16 -s 10 --verbose=false -ed "node_modules,.git"

  # 同时扫描文本和二进制文件 / Scan both text and binary files
  findx -t .txt,.log,.dll,.exe -f /path/to/scan

  # 使用所有简写参数 / Use all short flags
  findx -f /path/to/scan -t .txt,.log -k "password,token" -n 8 -s 10 -ed ".git" -ef "*.min.js"`
}

// GetUsageText 返回使用说明
func GetUsageText() string {
	return `findx [全局选项] / findx [global options]

参数说明 / Flag Description:
  简写和全称都可以使用 / Both short and long forms are available
  
  基础参数 / Basic Flags:
    -f, --folder      扫描目录（必填）
    -o, --output      输出文件路径
  
  文件类型 / File Types:
    -t, --type        指定文件类型
    -ta, --type-append 追加文件类型
  
  关键词 / Keywords:
    -k, --keyword     搜索关键词（二进制模式可为空）
    -ka, --keyword-append 追加关键词
  
  性能 / Performance:
    -n, --thread      线程数
    --verbose, --vb   实时输出
  
  高级 / Advanced:
    -s, --max-size    最大文件大小
    -ed, --exclude-dir 排除目录
    -ef, --exclude-file 排除文件
  
  二进制 / Binary:
    -b, --binary      二进制扫描模式
    --ctx, --context  上下文长度（字符数）

支持的文件类型 / Supported File Types:
  文本 / Text: .txt, .log, .ini, .conf, .yaml, .yml, .xml, .json, .sql, .properties, .md
  文档 / Document: .docx, .xlsx, .xls, .csv
  代码 / Code: .java, .py, .js, .php, .go, .c, .cpp, .h, .sh, .bat, .ps1
  二进制 / Binary: .dll, .exe, .so, .dylib, .bin, .o, .obj (PE文件敏感信息扫描)
  
注意 / Note:
  - 如果在 -t 或 -ta 中指定了二进制文件类型，会自动启用二进制扫描模式
  - If binary file types are specified in -t or -ta, binary scan mode will be enabled automatically

默认搜索关键词 / Default Keywords:
  password=, username=, jdbc:, user=, ssh-, ldap:, mysqli_connect,
  sk-, 账号, 密码, username:, password:`
}

// PrintBanner 打印Banner
func PrintBanner() {
	fmt.Println(Banner)
}
