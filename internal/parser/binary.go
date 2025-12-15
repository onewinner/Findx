package parser

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf16"

	"Findx/pkg/utils"
)

const (
	PE_SIGNATURE  = 0x00004550
	DOS_SIGNATURE = 0x5A4D
)

// DetectionRule 检测规则定义
type DetectionRule struct {
	Name        string
	Pattern     *regexp.Regexp
	Description string
	RiskLevel   string
}

// BinaryParser 二进制文件解析器（DLL/EXE）
type BinaryParser struct {
	rules []DetectionRule
}

// NewBinaryParser 创建二进制解析器
func NewBinaryParser() *BinaryParser {
	return &BinaryParser{
		rules: initDetectionRules(),
	}
}

// initDetectionRules 初始化检测规则
func initDetectionRules() []DetectionRule {
	return []DetectionRule{
		{
			Name:        "数据库连接字符串",
			Pattern:     regexp.MustCompile(`(?i)(ConnectionString|connstr?)\s*=\s*["']([^"']{10,200})["']`),
			Description: "数据库连接字符串包含认证信息",
			RiskLevel:   "high",
		},
		{
			Name:        "JDBC连接URL",
			Pattern:     regexp.MustCompile(`jdbc:\w+://[^\s"']+`),
			Description: "JDBC数据库连接URL",
			RiskLevel:   "high",
		},
		{
			Name:        "密码字段",
			Pattern:     regexp.MustCompile(`(?i)(password|pwd)\s*=\s*["']?([^"'\s;&]{4,50})`),
			Description: "密码字段赋值",
			RiskLevel:   "critical",
		},
		{
			Name:        "用户名字段",
			Pattern:     regexp.MustCompile(`(?i)(username|user|uid)\s*=\s*["']?([^"'\s;&]{3,50})`),
			Description: "用户名字段赋值",
			RiskLevel:   "high",
		},
		{
			Name:        "API密钥",
			Pattern:     regexp.MustCompile(`(?i)(api[_-]?key|apisecret|sk-)\s*=\s*["']?([a-zA-Z0-9]{20,60})`),
			Description: "API密钥或访问令牌",
			RiskLevel:   "critical",
		},
		{
			Name:        "SSH密钥",
			Pattern:     regexp.MustCompile(`ssh-\w+\s+[A-Za-z0-9+/]{100,}`),
			Description: "SSH公钥或私钥",
			RiskLevel:   "critical",
		},
		{
			Name:        "LDAP连接",
			Pattern:     regexp.MustCompile(`ldap[s]?://[^\s"']+`),
			Description: "LDAP连接字符串",
			RiskLevel:   "high",
		},
		{
			Name:        "MySQL连接",
			Pattern:     regexp.MustCompile(`mysqli_connect\([^)]+\)`),
			Description: "MySQL数据库连接",
			RiskLevel:   "high",
		},
		{
			Name:        "中文凭据",
			Pattern:     regexp.MustCompile(`(账号|密码|用户名)\s*[=:]\s*["']?([^"'\s]{3,50})`),
			Description: "中文账号密码信息",
			RiskLevel:   "high",
		},
		{
			Name:        "Bearer令牌",
			Pattern:     regexp.MustCompile(`Bearer\s+[\w\-._~+/]{20,100}`),
			Description: "Bearer认证令牌",
			RiskLevel:   "high",
		},
		{
			Name:        "私钥文件",
			Pattern:     regexp.MustCompile(`-----BEGIN (?:RSA|DSA|EC) PRIVATE KEY-----`),
			Description: "加密私钥文件",
			RiskLevel:   "critical",
		},
		{
			Name:        "邮箱地址",
			Pattern:     regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
			Description: "邮箱地址（可能用于认证）",
			RiskLevel:   "medium",
		},
		{
			Name:        "IP地址和端口",
			Pattern:     regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}(?::\d+)?\b`),
			Description: "IP地址和端口信息",
			RiskLevel:   "low",
		},
	}
}

// Parse 解析二进制文件内容
func (p *BinaryParser) Parse(filePath string, data []byte, verbose bool) []string {
	var matchingLines []string

	// 验证PE文件
	if len(data) < 64 || !isValidPEFile(data) {
		if verbose {
			fmt.Printf("[-] 不是有效的PE文件: %s\n", filePath)
		}
		return matchingLines
	}

	if verbose {
		fmt.Printf("[*] 分析二进制文件: %s (%.2f MB)\n", filePath, float64(len(data))/1024/1024)
	}

	// 提取字符串
	allStrings := extractMeaningfulStrings(data)

	// 检查字符串
	for _, str := range allStrings {
		results := p.checkStringWithRules(str, data)
		for _, result := range results {
			lineOutput := fmt.Sprintf("[+] %s: %s", result.RuleName, utils.TruncateString(result.MatchedValue, 100))
			matchingLines = append(matchingLines, lineOutput)
			if verbose {
				fmt.Println(lineOutput)
			}
		}
	}

	// 检查Base64编码
	base64Results := p.checkBase64Encoded(data)
	for _, result := range base64Results {
		lineOutput := fmt.Sprintf("[+] %s (Base64): %s", result.RuleName, utils.TruncateString(result.MatchedValue, 100))
		matchingLines = append(matchingLines, lineOutput)
		if verbose {
			fmt.Println(lineOutput)
		}
	}

	return matchingLines
}

// ParseWithKeywords 使用关键字解析二进制文件内容
func (p *BinaryParser) ParseWithKeywords(filePath string, data []byte, keywords []string, verbose bool, contextLen int) []string {
	var matchingLines []string
	seenOffsets := make(map[int]bool) // 用于去重

	// 验证PE文件
	if len(data) < 64 || !isValidPEFile(data) {
		if verbose {
			fmt.Printf("[-] 不是有效的PE文件: %s\n", filePath)
		}
		return matchingLines
	}

	if verbose {
		fmt.Printf("[*] 分析二进制文件: %s (%.2f MB)\n", filePath, float64(len(data))/1024/1024)
	}

	// 提取字符串
	allStrings := extractMeaningfulStrings(data)

	// 1. 使用规则检查
	for _, str := range allStrings {
		results := p.checkStringWithRulesEx(str, data, contextLen)
		for _, result := range results {
			// 去重：检查偏移是否已存在
			if seenOffsets[result.Offset] {
				continue
			}
			seenOffsets[result.Offset] = true
			
			lineOutput := formatBinaryResult(result, "规则匹配", contextLen)
			matchingLines = append(matchingLines, lineOutput)
			if verbose {
				fmt.Println(lineOutput)
			}
		}
	}

	// 2. 使用关键字检查
	if len(keywords) > 0 {
		for _, str := range allStrings {
			for _, keyword := range keywords {
				if strings.Contains(str, keyword) {
					offset := findStringOffset(data, str)
					
					// 去重：检查偏移是否已存在
					if seenOffsets[offset] {
						continue
					}
					seenOffsets[offset] = true
					
					context := getStringContext(data, offset, contextLen)
					
					result := BinaryMatchResult{
						RuleName:     "关键字匹配",
						RuleDesc:     fmt.Sprintf("匹配关键字: %s", keyword),
						RiskLevel:    "medium",
						MatchedValue: str,
						Offset:       offset,
						Context:      context,
					}
					
					lineOutput := formatBinaryResult(result, "关键字", contextLen)
					matchingLines = append(matchingLines, lineOutput)
					if verbose {
						fmt.Println(lineOutput)
					}
					break // 找到一个匹配即可
				}
			}
		}
	}

	// 3. 检查Base64编码
	base64Results := p.checkBase64EncodedEx(data, contextLen)
	for _, result := range base64Results {
		// 去重：检查偏移是否已存在
		if seenOffsets[result.Offset] {
			continue
		}
		seenOffsets[result.Offset] = true
		
		lineOutput := formatBinaryResult(result, "Base64编码", contextLen)
		matchingLines = append(matchingLines, lineOutput)
		if verbose {
			fmt.Println(lineOutput)
		}
	}

	return matchingLines
}

// formatBinaryResult 格式化二进制扫描结果
func formatBinaryResult(result BinaryMatchResult, matchType string, contextLen int) string {
	// 根据上下文长度动态调整显示
	contextDisplay := result.Context
	if len(contextDisplay) > contextLen {
		contextDisplay = contextDisplay[:contextLen] + "..."
	}
	
	return fmt.Sprintf("BINARY|%s|%s|%s|%s|0x%X|%s",
		matchType,
		result.RuleName,
		result.RiskLevel,
		result.MatchedValue,
		result.Offset,
		contextDisplay)
}



// checkStringWithRulesEx 使用规则检查字符串（支持自定义上下文长度）
func (p *BinaryParser) checkStringWithRulesEx(str string, data []byte, contextLen int) []BinaryMatchResult {
	var results []BinaryMatchResult

	for _, rule := range p.rules {
		matches := rule.Pattern.FindAllStringSubmatch(str, -1)
		for _, match := range matches {
			if len(match) > 1 {
				matchedValue := match[1]
				if len(match) > 2 {
					matchedValue = match[2]
				}

				if isValidCredential(matchedValue) {
					// 尝试多种方式查找偏移
					offset := findStringOffset(data, match[0])
					if offset == -1 {
						offset = findStringOffset(data, matchedValue)
					}
					if offset == -1 {
						offset = findStringOffset(data, str)
					}
					
					// 如果还是找不到，尝试查找部分字符串
					if offset == -1 && len(matchedValue) > 10 {
						offset = findStringOffset(data, matchedValue[:10])
					}

					context := getStringContext(data, offset, contextLen)
					
					// 如果找不到偏移，使用原始字符串作为上下文
					if offset == -1 && context == "无法定位" {
						context = truncateForContext(str, contextLen)
					}

					result := BinaryMatchResult{
						RuleName:     rule.Name,
						RuleDesc:     rule.Description,
						RiskLevel:    rule.RiskLevel,
						MatchedValue: matchedValue,
						Offset:       offset,
						Context:      context,
					}
					results = append(results, result)
				}
			}
		}
	}

	return results
}

// truncateForContext 截断字符串用作上下文
func truncateForContext(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	half := maxLen / 2
	return s[:half] + "..." + s[len(s)-half:]
}

// checkBase64EncodedEx 检查Base64编码的内容（支持自定义上下文长度）
func (p *BinaryParser) checkBase64EncodedEx(data []byte, contextLen int) []BinaryMatchResult {
	var results []BinaryMatchResult
	base64Pattern := regexp.MustCompile(`[A-Za-z0-9+/]{40,}[=]{0,2}`)

	base64Matches := base64Pattern.FindAllIndex(data, -1)
	for _, match := range base64Matches {
		start, end := match[0], match[1]
		base64Str := string(data[start:end])

		// 尝试解码
		decoded, err := base64.StdEncoding.DecodeString(base64Str)
		if err != nil {
			continue
		}

		// 检查解码后是否为文本
		if !isText(decoded) {
			continue
		}

		decodedStr := string(decoded)

		// 对解码后的文本应用所有检测规则
		for _, rule := range p.rules {
			ruleMatches := rule.Pattern.FindAllStringSubmatch(decodedStr, -1)
			for _, ruleMatch := range ruleMatches {
				if len(ruleMatch) < 2 {
					continue
				}

				matchedValue := ruleMatch[1]
				if len(ruleMatch) > 2 {
					matchedValue = ruleMatch[2]
				}

				if !isValidCredential(matchedValue) {
					continue
				}

				context := getStringContext(data, start, contextLen)
				
				// 如果上下文无法定位，使用解码后的字符串
				if context == "无法定位" {
					context = fmt.Sprintf("Base64: %s -> %s", 
						truncateForContext(base64Str, contextLen/2),
						truncateForContext(decodedStr, contextLen/2))
				}

				results = append(results, BinaryMatchResult{
					RuleName:     rule.Name + " (Base64编码)",
					RuleDesc:     rule.Description + " - Base64编码版本",
					RiskLevel:    rule.RiskLevel,
					MatchedValue: matchedValue,
					Offset:       start,
					Context:      context,
				})
			}
		}
	}

	return results
}

// BinaryMatchResult 二进制匹配结果
type BinaryMatchResult struct {
	RuleName     string
	RuleDesc     string
	RiskLevel    string
	MatchedValue string
	Offset       int
	Context      string
}

// checkStringWithRules 使用规则检查字符串
func (p *BinaryParser) checkStringWithRules(str string, data []byte) []BinaryMatchResult {
	var results []BinaryMatchResult

	for _, rule := range p.rules {
		matches := rule.Pattern.FindAllStringSubmatch(str, -1)
		for _, match := range matches {
			if len(match) > 1 {
				matchedValue := match[1]
				if len(match) > 2 {
					matchedValue = match[2]
				}

				if isValidCredential(matchedValue) {
					offset := findStringOffset(data, match[0])
					if offset == -1 {
						offset = findStringOffset(data, matchedValue)
					}

					context := getStringContext(data, offset, 50)

					result := BinaryMatchResult{
						RuleName:     rule.Name,
						RuleDesc:     rule.Description,
						RiskLevel:    rule.RiskLevel,
						MatchedValue: matchedValue,
						Offset:       offset,
						Context:      context,
					}
					results = append(results, result)
				}
			}
		}
	}

	return results
}

// checkBase64Encoded 检查Base64编码的内容
func (p *BinaryParser) checkBase64Encoded(data []byte) []BinaryMatchResult {
	var results []BinaryMatchResult
	base64Pattern := regexp.MustCompile(`[A-Za-z0-9+/]{40,}[=]{0,2}`)

	base64Matches := base64Pattern.FindAllIndex(data, -1)
	for _, match := range base64Matches {
		start, end := match[0], match[1]
		base64Str := string(data[start:end])

		// 尝试解码
		decoded, err := base64.StdEncoding.DecodeString(base64Str)
		if err != nil {
			continue
		}

		// 检查解码后是否为文本
		if !isText(decoded) {
			continue
		}

		decodedStr := string(decoded)

		// 对解码后的文本应用所有检测规则
		for _, rule := range p.rules {
			ruleMatches := rule.Pattern.FindAllStringSubmatch(decodedStr, -1)
			for _, ruleMatch := range ruleMatches {
				if len(ruleMatch) < 2 {
					continue
				}

				matchedValue := ruleMatch[1]
				if len(ruleMatch) > 2 {
					matchedValue = ruleMatch[2]
				}

				if !isValidCredential(matchedValue) {
					continue
				}

				context := getStringContext(data, start, 50)

				results = append(results, BinaryMatchResult{
					RuleName:     rule.Name + " (Base64编码)",
					RuleDesc:     rule.Description + " - Base64编码版本",
					RiskLevel:    rule.RiskLevel,
					MatchedValue: matchedValue,
					Offset:       start,
					Context:      context,
				})
			}
		}
	}

	return results
}

// extractMeaningfulStrings 提取有意义的字符串
func extractMeaningfulStrings(data []byte) []string {
	var results []string
	stringSet := make(map[string]bool)

	// 提取UTF-8字符串
	var current strings.Builder
	for i := 0; i < len(data); i++ {
		if data[i] >= 32 && data[i] <= 126 {
			current.WriteByte(data[i])
		} else {
			if current.Len() >= 8 {
				str := current.String()
				if !stringSet[str] && isMeaningfulString(str) {
					stringSet[str] = true
					results = append(results, str)
				}
			}
			current.Reset()
		}
	}

	if current.Len() >= 8 {
		str := current.String()
		if !stringSet[str] && isMeaningfulString(str) {
			results = append(results, str)
		}
	}

	// 提取UTF-16字符串
	utf16Strings := extractUTF16Strings(data)
	for _, str := range utf16Strings {
		if !stringSet[str] && isMeaningfulString(str) {
			stringSet[str] = true
			results = append(results, str)
		}
	}

	return results
}

// extractUTF16Strings 提取UTF-16字符串
func extractUTF16Strings(data []byte) []string {
	var results []string
	var currentString []uint16

	for i := 0; i < len(data)-1; i += 2 {
		char := binary.LittleEndian.Uint16(data[i:])
		if char >= 32 && char <= 126 {
			currentString = append(currentString, char)
		} else {
			if len(currentString) >= 8 {
				str := string(utf16.Decode(currentString))
				if isMeaningfulString(str) {
					results = append(results, str)
				}
			}
			currentString = nil
		}
	}

	if len(currentString) >= 8 {
		str := string(utf16.Decode(currentString))
		if isMeaningfulString(str) {
			results = append(results, str)
		}
	}

	return results
}

// isMeaningfulString 判断字符串是否有意义
func isMeaningfulString(str string) bool {
	if len(str) > 500 || len(str) < 8 {
		return false
	}

	if isLikelyGarbage(str) {
		return false
	}

	keywords := []string{
		"password", "user", "jdbc", "ssh", "ldap", "mysql", "sk-",
		"账号", "密码", "Connection", "Database", "Server", "Host",
		"Token", "Key", "Secret", "Auth", "Login", "credential",
		"connect", "config", "setting", "account", "passwd",
	}

	for _, keyword := range keywords {
		if strings.Contains(strings.ToLower(str), strings.ToLower(keyword)) {
			return true
		}
	}

	if strings.Contains(str, "://") || strings.Contains(str, "Data Source") ||
		strings.Contains(str, "Initial Catalog") || strings.Contains(str, "User ID") {
		return true
	}

	return false
}

// isLikelyGarbage 判断是否为垃圾字符串
func isLikelyGarbage(str string) bool {
	if len(str) < 10 {
		return false
	}

	specialCharCount := 0
	for _, ch := range str {
		if (ch < 32 || ch > 126) && ch != ' ' && ch != '\t' && ch != '\n' {
			specialCharCount++
		}
	}

	if float64(specialCharCount)/float64(len(str)) > 0.3 {
		return true
	}

	if hasRepeatingPattern(str) {
		return true
	}

	return false
}

// hasRepeatingPattern 检查是否有重复模式
func hasRepeatingPattern(str string) bool {
	if len(str) < 20 {
		return false
	}

	for i := 0; i < len(str)-10; i++ {
		if str[i] == str[i+1] && str[i] == str[i+2] && str[i] == str[i+3] {
			return true
		}
	}

	return false
}

// isValidCredential 验证是否为有效凭据
func isValidCredential(str string) bool {
	if len(str) < 3 {
		return false
	}

	excluded := []string{
		"true", "false", "null", "void", "main", "class", "string", "int", "bool",
		"==", "!=", "=<", "=>", "= ", " =", "=p=", "=6", "=N", "=L", "=W", "=f", "=C",
		"GET", "POST", "HTTP", "Content", "Type", "Length", "Version", "Microsoft",
		"System", "Windows", "Assembly", "PublicKey", "Culture", "Token",
	}

	for _, exclude := range excluded {
		if strings.Contains(str, exclude) {
			return false
		}
	}

	return hasCredentialLikePattern(str)
}

// hasCredentialLikePattern 检查是否符合凭据模式
func hasCredentialLikePattern(str string) bool {
	validPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+-=]{4,50}$`),
		regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`),
		regexp.MustCompile(`^[a-zA-Z0-9]{20,60}$`),
		regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]{3,20}$`),
		regexp.MustCompile(`^[a-zA-Z]+://`),
		regexp.MustCompile(`^jdbc:\w+://`),
	}

	for _, pattern := range validPatterns {
		if pattern.MatchString(str) {
			return true
		}
	}

	return false
}

// isText 判断是否为文本
func isText(data []byte) bool {
	printable := 0
	for _, b := range data {
		if b >= 32 && b <= 126 || b == 10 || b == 13 || b == 9 {
			printable++
		}
	}
	return float64(printable)/float64(len(data)) > 0.7
}

// isValidPEFile 验证是否为有效的PE文件
func isValidPEFile(data []byte) bool {
	if len(data) < 2 || binary.LittleEndian.Uint16(data[0:2]) != DOS_SIGNATURE {
		return false
	}
	if len(data) < 0x40 {
		return false
	}
	peOffset := int(binary.LittleEndian.Uint32(data[0x3C:0x40]))
	return peOffset+4 <= len(data) &&
		binary.LittleEndian.Uint32(data[peOffset:peOffset+4]) == PE_SIGNATURE
}

// findStringOffset 查找字符串在数据中的偏移
func findStringOffset(data []byte, str string) int {
	return strings.Index(string(data), str)
}

// getStringContext 获取字符串上下文
func getStringContext(data []byte, offset, contextLen int) string {
	if offset == -1 {
		return "无法定位"
	}

	start := max(0, offset-contextLen)
	end := min(len(data), offset+contextLen)

	contextBytes := data[start:end]
	var result strings.Builder
	for _, b := range contextBytes {
		if b >= 32 && b <= 126 {
			result.WriteByte(b)
		} else {
			result.WriteByte('.')
		}
	}

	return result.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
