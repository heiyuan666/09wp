package service

import (
	"PanCheck/internal/model"
	"PanCheck/internal/repository"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/474420502/gcurl"
	"github.com/dop251/goja"
)

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ScheduledTaskService 任务计划管理服务
type ScheduledTaskService struct {
	taskRepo      *repository.ScheduledTaskRepository
	executionRepo *repository.TaskExecutionRepository
}

// NewScheduledTaskService 创建任务计划管理服务
func NewScheduledTaskService() *ScheduledTaskService {
	return &ScheduledTaskService{
		taskRepo:      repository.NewScheduledTaskRepository(),
		executionRepo: repository.NewTaskExecutionRepository(),
	}
}

// CreateTask 创建任务计划
func (s *ScheduledTaskService) CreateTask(task *model.ScheduledTask) error {
	// 检查任务名称是否已存在
	exists, err := s.taskRepo.ExistsByName(task.Name, nil)
	if err != nil {
		return fmt.Errorf("检查任务名称失败: %v", err)
	}
	if exists {
		return fmt.Errorf("任务名称 '%s' 已存在，请使用其他名称", task.Name)
	}

	// 设置默认状态
	if task.Status == "" {
		task.Status = "stopped"
	}

	return s.taskRepo.Create(task)
}

// GetTask 获取任务计划
func (s *ScheduledTaskService) GetTask(id uint) (*model.ScheduledTask, error) {
	return s.taskRepo.GetByID(id)
}

// UpdateTask 更新任务计划
func (s *ScheduledTaskService) UpdateTask(task *model.ScheduledTask) error {
	// 如果更新了任务名称，检查新名称是否已被其他任务使用
	if task.Name != "" {
		excludeID := &task.ID
		exists, err := s.taskRepo.ExistsByName(task.Name, excludeID)
		if err != nil {
			return fmt.Errorf("检查任务名称失败: %v", err)
		}
		if exists {
			return fmt.Errorf("任务名称 '%s' 已存在，请使用其他名称", task.Name)
		}
	}

	return s.taskRepo.Update(task)
}

// DeleteTask 删除任务计划
func (s *ScheduledTaskService) DeleteTask(id uint) error {
	return s.taskRepo.Delete(id)
}

// ListTasks 分页查询任务计划
func (s *ScheduledTaskService) ListTasks(page, pageSize int, tags []string, status string) ([]model.ScheduledTask, int64, error) {
	return s.taskRepo.List(page, pageSize, tags, status)
}

// GetAllTags 获取所有标签列表
func (s *ScheduledTaskService) GetAllTags() ([]string, error) {
	return s.taskRepo.GetAllTags()
}

// EnableTask 启用任务
func (s *ScheduledTaskService) EnableTask(id uint) error {
	task, err := s.taskRepo.GetByID(id)
	if err != nil {
		return err
	}
	task.Status = "active"
	return s.taskRepo.Update(task)
}

// DisableTask 禁用任务
func (s *ScheduledTaskService) DisableTask(id uint) error {
	task, err := s.taskRepo.GetByID(id)
	if err != nil {
		return err
	}
	task.Status = "stopped"
	return s.taskRepo.Update(task)
}

// CheckExpiredTasks 检查并更新过期任务
func (s *ScheduledTaskService) CheckExpiredTasks() error {
	expiredTasks, err := s.taskRepo.GetExpiredTasks()
	if err != nil {
		return err
	}

	for _, task := range expiredTasks {
		task.Status = "expired"
		if err := s.taskRepo.Update(&task); err != nil {
			return err
		}
	}

	return nil
}

// normalizeCurlCommand 将Postman格式的curl命令转换为标准格式（供gcurl使用）
func normalizeCurlCommand(curlCommand string) string {
	// 先处理多行命令：合并续行
	lines := strings.Split(curlCommand, "\n")
	var mergedLines []string
	var currentLine strings.Builder

	for i, line := range lines {
		trimmedLine := strings.TrimRight(line, " \t\r")
		hasContinuation := strings.HasSuffix(trimmedLine, "\\")

		// 移除续行符
		if hasContinuation {
			trimmedLine = strings.TrimSuffix(trimmedLine, "\\")
		}

		// 添加当前行内容（去除首尾空格）
		trimmedLine = strings.TrimSpace(trimmedLine)
		if trimmedLine != "" {
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(trimmedLine)
		}

		// 如果没有续行符，完成当前命令
		if !hasContinuation || i == len(lines)-1 {
			if currentLine.Len() > 0 {
				mergedLines = append(mergedLines, currentLine.String())
				currentLine.Reset()
			}
		}
	}

	// 合并所有行
	mergedCommand := strings.Join(mergedLines, " ")

	// 替换Postman格式为标准格式（gcurl 支持标准 curl 命令格式）
	mergedCommand = strings.ReplaceAll(mergedCommand, "--request", "-X")
	mergedCommand = strings.ReplaceAll(mergedCommand, "--url", "")
	mergedCommand = strings.ReplaceAll(mergedCommand, "--header", "-H")
	mergedCommand = strings.ReplaceAll(mergedCommand, "--data", "-d")
	mergedCommand = strings.ReplaceAll(mergedCommand, "--data-raw", "-d")
	mergedCommand = strings.ReplaceAll(mergedCommand, "--data-binary", "--data-binary")
	mergedCommand = strings.ReplaceAll(mergedCommand, "--cookie", "-b")
	mergedCommand = strings.ReplaceAll(mergedCommand, "--user", "-u")

	// 移除多余的空格
	result := strings.ReplaceAll(mergedCommand, "  ", " ")
	result = strings.TrimSpace(result)

	return result
}

// ExecuteCurlCommand 执行curl命令（使用gcurl库）
func (s *ScheduledTaskService) ExecuteCurlCommand(curlCommand string) (string, error) {
	// 清理命令：移除换行符和多余空格
	// curlCommand = strings.TrimSpace(curlCommand)
	// // 替换转义字符
	// curlCommand = strings.ReplaceAll(curlCommand, "\\n", "\n")
	// curlCommand = strings.ReplaceAll(curlCommand, "\\t", "\t")

	// // 标准化curl命令格式（将Postman格式转换为标准格式）
	// curlCommand = normalizeCurlCommand(curlCommand)

	// log.Printf("[ExecuteCurlCommand] normalized curl command: %s", curlCommand)

	// 使用 gcurl 解析 curl 命令
	curl, err := gcurl.Parse(curlCommand)
	if err != nil {
		log.Printf("[ExecuteCurlCommand] failed to parse curl command: %v", err)
		return "", fmt.Errorf("failed to parse curl command: %v", err)
	}

	// 执行 HTTP 请求
	resp, err := curl.Request().Execute()
	if err != nil {
		log.Printf("[ExecuteCurlCommand] HTTP request failed: %v", err)
		// 检查响应是否包含有效数据
		if resp != nil {
			contentBytes := resp.Content()
			if len(contentBytes) > 0 {
				outputStr := string(contentBytes)
				hasValidData := strings.Contains(outputStr, "http://") ||
					strings.Contains(outputStr, "https://") ||
					strings.Contains(outputStr, "magnet:")
				if hasValidData {
					log.Printf("[ExecuteCurlCommand] request had error but response contains valid data, returning data")
					return outputStr, nil
				}
			}
		}
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}

	// 获取响应内容（Content() 返回 []byte，需要转换为 string）
	contentBytes := resp.Content()
	outputStr := string(contentBytes)

	// // 打印日志：curl 执行结果
	// log.Printf("[ExecuteCurlCommand] HTTP request executed successfully")
	// // 尝试获取状态码（如果响应对象有 StatusCode 字段）
	// if resp != nil {
	// 	log.Printf("[ExecuteCurlCommand] response received, content length: %d bytes", len(contentBytes))
	// }
	// log.Printf("[ExecuteCurlCommand] output length: %d bytes", len(outputStr))
	// if len(outputStr) > 0 {
	// 	// 只打印前500个字符，避免日志过长
	// 	preview := outputStr
	// 	if len(preview) > 500 {
	// 		preview = preview[:500] + "..."
	// 	}
	// 	log.Printf("[ExecuteCurlCommand] output preview: %s", preview)
	// }

	// // 清理响应内容，提取 JSON 数据（如果需要）
	// outputStr = cleanCurlOutput(outputStr)
	// log.Printf("[ExecuteCurlCommand] cleaned output length: %d bytes", len(outputStr))
	// if len(outputStr) > 0 && len(outputStr) <= 500 {
	// 	log.Printf("[ExecuteCurlCommand] cleaned output: %s", outputStr)
	// } else if len(outputStr) > 500 {
	// 	log.Printf("[ExecuteCurlCommand] cleaned output preview: %s...", outputStr[:500])
	// }

	return outputStr, nil
}

// cleanCurlOutput 清理 curl 输出，提取 JSON 数据
// 由于使用了 -s 参数，curl 不会输出进度信息，所以只需要提取 JSON 部分
func cleanCurlOutput(output string) string {
	log.Printf("[cleanCurlOutput] input length: %d bytes", len(output))

	// 如果输出为空，直接返回
	if len(output) == 0 {
		log.Printf("[cleanCurlOutput] output is empty")
		return output
	}

	// 查找第一个 { 或 [ 的位置（JSON 开始）
	jsonStartIdx := strings.Index(output, "{")
	isArray := false
	if jsonStartIdx == -1 {
		jsonStartIdx = strings.Index(output, "[")
		isArray = true
	}

	log.Printf("[cleanCurlOutput] JSON start index: %d, isArray: %v", jsonStartIdx, isArray)

	if jsonStartIdx >= 0 {
		// 找到 JSON 开始位置，提取 JSON 部分
		jsonPart := output[jsonStartIdx:]
		log.Printf("[cleanCurlOutput] extracted JSON part, length: %d", len(jsonPart))

		// 使用更智能的方法提取完整的 JSON：通过匹配括号来找到正确的结束位置
		jsonEndIdx := findMatchingBrace(jsonPart, isArray)
		if jsonEndIdx > 0 {
			jsonPart = jsonPart[:jsonEndIdx+1]
			log.Printf("[cleanCurlOutput] found matching brace at index: %d", jsonEndIdx)
		} else {
			// 如果找不到匹配的括号，回退到原来的方法
			log.Printf("[cleanCurlOutput] failed to find matching brace, using fallback method")
			lastBraceIdx := strings.LastIndex(jsonPart, "}")
			lastBracketIdx := strings.LastIndex(jsonPart, "]")
			jsonEndIdx = lastBraceIdx
			if lastBracketIdx > lastBraceIdx {
				jsonEndIdx = lastBracketIdx
			}
			if jsonEndIdx > 0 {
				jsonPart = jsonPart[:jsonEndIdx+1]
			}
		}

		log.Printf("[cleanCurlOutput] extracted JSON from curl output, original length: %d, JSON length: %d", len(output), len(jsonPart))
		cleaned := strings.TrimSpace(jsonPart)

		// 清理 Windows 换行符 \r，避免 JSON 解析错误
		cleaned = strings.ReplaceAll(cleaned, "\r\n", "\n")
		cleaned = strings.ReplaceAll(cleaned, "\r", "")

		// 修复 JSON 字符串值中的未转义换行符（处理 API 返回的 JSON 中可能包含的未转义换行符）
		cleaned = fixUnescapedNewlinesInJSON(cleaned)

		// 验证是否是有效的 JSON
		var test interface{}
		if err := json.Unmarshal([]byte(cleaned), &test); err != nil {
			log.Printf("[cleanCurlOutput] WARNING: extracted JSON is invalid after cleaning: %v", err)
			log.Printf("[cleanCurlOutput] cleaned JSON preview (first 200 chars): %s", cleaned[:min(200, len(cleaned))])
			log.Printf("[cleanCurlOutput] cleaned JSON ending (last 200 chars): %s", cleaned[max(0, len(cleaned)-200):])

			// 尝试找到错误位置附近的字符，帮助调试
			if jsonErr, ok := err.(*json.SyntaxError); ok {
				offset := int(jsonErr.Offset)
				start := max(0, offset-50)
				end := min(len(cleaned), offset+50)
				log.Printf("[cleanCurlOutput] JSON syntax error at offset %d (out of %d), context: %q", offset, len(cleaned), cleaned[start:end])

				// 如果错误位置接近末尾，检查是否有额外的字符
				if offset > len(cleaned)-10 {
					log.Printf("[cleanCurlOutput] Error near end of JSON, checking for extra characters")
					// 显示末尾的字符（包括不可见字符）
					endPart := cleaned[max(0, offset-20):]
					log.Printf("[cleanCurlOutput] End part (with hex): %q", endPart)
				}
			}

			// 尝试修复常见的 JSON 问题：移除 JSON 后面的额外字符
			// 如果 JSON 后面有非空白字符，尝试只保留到最后一个有效的 }
			cleanedTrimmed := strings.TrimRight(cleaned, " \t\n\r")
			if len(cleanedTrimmed) < len(cleaned) {
				// 有尾部空白，尝试重新验证
				if err2 := json.Unmarshal([]byte(cleanedTrimmed), &test); err2 == nil {
					log.Printf("[cleanCurlOutput] JSON is valid after trimming trailing whitespace")
					return cleanedTrimmed
				}
			}

			// 尝试找到最后一个有效的 JSON 结束位置
			// 从后往前查找，找到第一个能成功解析的位置
			for i := len(cleanedTrimmed) - 1; i > 0; i-- {
				if cleanedTrimmed[i] == '}' || cleanedTrimmed[i] == ']' {
					testPart := cleanedTrimmed[:i+1]
					if err2 := json.Unmarshal([]byte(testPart), &test); err2 == nil {
						log.Printf("[cleanCurlOutput] found valid JSON by truncating at position %d", i+1)
						return testPart
					}
				}
			}

			// 即使验证失败，也返回清理后的 JSON 部分
			// 让 JavaScript 脚本自己处理可能的格式问题
			log.Printf("[cleanCurlOutput] returning cleaned JSON despite validation error (may work in JS)")
			return cleaned
		}
		log.Printf("[cleanCurlOutput] successfully extracted valid JSON, cleaned length: %d", len(cleaned))
		if len(cleaned) <= 500 {
			log.Printf("[cleanCurlOutput] cleaned JSON: %s", cleaned)
		} else {
			log.Printf("[cleanCurlOutput] cleaned JSON preview: %s...", cleaned[:500])
		}
		return cleaned
	}

	// 如果没有找到 JSON，返回清理后的原始输出
	log.Printf("[cleanCurlOutput] no JSON found in output, returning cleaned output")
	return strings.TrimSpace(output)
}

// findMatchingBrace 找到匹配的括号位置（支持 { } 和 [ ]）
func findMatchingBrace(jsonStr string, isArray bool) int {
	openChar := byte('{')
	closeChar := byte('}')
	if isArray {
		openChar = byte('[')
		closeChar = byte(']')
	}

	depth := 0
	inString := false
	escapeNext := false

	for i := 0; i < len(jsonStr); i++ {
		char := jsonStr[i]

		if escapeNext {
			escapeNext = false
			continue
		}

		if char == '\\' {
			escapeNext = true
			continue
		}

		if char == '"' {
			inString = !inString
			continue
		}

		if !inString {
			if char == openChar {
				depth++
			} else if char == closeChar {
				depth--
				if depth == 0 {
					return i
				}
			}
		}
	}

	return -1
}

// fixUnescapedNewlinesInJSON 修复 JSON 字符串值中的未转义换行符
// 将字符串值中的实际换行符转义为 \n
func fixUnescapedNewlinesInJSON(jsonStr string) string {
	// 使用正则表达式匹配 JSON 字符串值中的未转义换行符
	// 匹配模式：在双引号内的内容，但不包括已经转义的 \n
	// 这个正则表达式匹配：双引号开始 -> 任意字符（包括换行符）-> 双引号结束
	// 但需要避免匹配已经转义的 \n

	// 更简单的方法：使用状态机或者分步处理
	// 1. 先找到所有字符串值的范围
	// 2. 在这些范围内替换未转义的换行符

	var result strings.Builder
	result.Grow(len(jsonStr))

	inString := false
	escapeNext := false

	for i := 0; i < len(jsonStr); i++ {
		char := jsonStr[i]

		if escapeNext {
			// 当前字符是转义字符后的字符，直接添加
			result.WriteByte(char)
			escapeNext = false
			continue
		}

		if char == '\\' {
			// 转义字符
			result.WriteByte(char)
			escapeNext = true
			continue
		}

		if char == '"' {
			// 双引号，切换字符串状态
			result.WriteByte(char)
			inString = !inString
			continue
		}

		if inString {
			// 在字符串值内
			if char == '\n' {
				// 未转义的换行符，需要转义
				result.WriteString("\\n")
			} else if char == '\t' {
				// 未转义的制表符，需要转义
				result.WriteString("\\t")
			} else if char == '\r' {
				// 未转义的回车符，忽略（已经在之前清理过了）
				continue
			} else {
				result.WriteByte(char)
			}
		} else {
			// 在字符串值外，直接添加
			result.WriteByte(char)
		}
	}

	return result.String()
}

// TransformData 执行JavaScript转换脚本
func (s *ScheduledTaskService) TransformData(rawData string, script string) ([]string, error) {
	// 打印日志：输入数据
	log.Printf("[TransformData] rawData length: %d bytes", len(rawData))
	if len(rawData) > 0 {
		preview := rawData
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		log.Printf("[TransformData] rawData preview: %s", preview)
	}
	log.Printf("[TransformData] script length: %d bytes", len(script))
	if len(script) > 0 {
		preview := script
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		log.Printf("[TransformData] script preview: %s", preview)
	}

	// 清理脚本：移除转义字符
	script = strings.TrimSpace(script)
	// 注意：不要替换脚本中的\n，因为脚本中可能需要使用\n作为字符串字面量
	// 只在需要时替换实际的换行符

	if script == "" || script == "return rawData;" {
		log.Printf("[TransformData] no script or default script, trying to parse rawData directly")
		// 如果没有转换脚本或只是返回原始数据，尝试多种方式解析
		// 1. 尝试解析为JSON数组
		var jsonResult []string
		if err := json.Unmarshal([]byte(rawData), &jsonResult); err == nil && len(jsonResult) > 0 {
			log.Printf("[TransformData] successfully parsed as JSON array, count: %d", len(jsonResult))
			return jsonResult, nil
		} else {
			log.Printf("[TransformData] failed to parse as JSON array: %v", err)
		}

		// 2. 尝试按行分割（去除空行和curl错误信息）
		lines := strings.Split(rawData, "\n")
		result := make([]string, 0)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// 过滤掉curl的错误信息、进度信息和空行
			if line != "" &&
				!strings.HasPrefix(line, "curl:") &&
				!strings.HasPrefix(line, "  %") &&
				!strings.HasPrefix(line, "% Total") &&
				!strings.HasPrefix(line, "Dload") &&
				!strings.HasPrefix(line, "Upload") &&
				!strings.HasPrefix(line, "Total") &&
				!strings.HasPrefix(line, "Spent") &&
				!strings.HasPrefix(line, "Left") &&
				!strings.HasPrefix(line, "Speed") &&
				!strings.HasPrefix(line, "--:--:--") &&
				!strings.HasPrefix(line, "100") {
				result = append(result, line)
			}
		}
		if len(result) > 0 {
			log.Printf("[TransformData] successfully parsed as lines, count: %d", len(result))
			return result, nil
		}

		log.Printf("[TransformData] failed to parse raw data: not a valid JSON array and no valid lines found")
		return nil, fmt.Errorf("failed to parse raw data: not a valid JSON array and no valid lines found")
	}

	// 使用goja执行JavaScript代码
	vm := goja.New()

	// 设置输入数据
	log.Printf("[TransformData] setting rawData in VM, length: %d", len(rawData))
	if err := vm.Set("rawData", rawData); err != nil {
		log.Printf("[TransformData] failed to set rawData: %v", err)
		return nil, fmt.Errorf("failed to set rawData: %v", err)
	}

	// 验证 rawData 是否是有效的 JSON（如果脚本需要解析）
	// 尝试解析 JSON 以验证格式
	var testJSON interface{}
	if err := json.Unmarshal([]byte(rawData), &testJSON); err != nil {
		log.Printf("[TransformData] WARNING: rawData is not valid JSON: %v", err)
		log.Printf("[TransformData] rawData first 200 chars: %s", rawData[:min(200, len(rawData))])
	} else {
		log.Printf("[TransformData] rawData is valid JSON, type: %T", testJSON)
	}

	// 处理脚本中的转义字符：将 \\n 替换为 \n（在JavaScript字符串中）
	// 但要注意，如果脚本中已经有正确的\n，不要重复替换
	script = strings.ReplaceAll(script, "\\\\n", "\n")
	script = strings.ReplaceAll(script, "\\\\t", "\t")

	// 包装脚本，确保在函数作用域内执行，这样可以支持return语句
	// 检查脚本是否已经是完整的函数表达式（以 (function 或 (() => 开头）
	isFunctionExpression := strings.HasPrefix(strings.TrimSpace(script), "(function") ||
		strings.HasPrefix(strings.TrimSpace(script), "(() =>") ||
		strings.HasPrefix(strings.TrimSpace(script), "(async function")

	wrappedScript := script
	if !isFunctionExpression {
		// 如果脚本包含return语句，需要包装在函数中
		if strings.Contains(script, "return") {
			wrappedScript = "(function() { " + script + " })()"
		} else {
			// 如果脚本没有return，也包装在函数中，最后返回结果
			wrappedScript = "(function() { " + script + "; return result || urlArray || []; })()"
		}
	}

	// 打印包装后的脚本
	log.Printf("[TransformData] wrapped script length: %d bytes", len(wrappedScript))
	if len(wrappedScript) > 1000 {
		log.Printf("[TransformData] wrapped script preview: %s...", wrappedScript[:1000])
	} else {
		log.Printf("[TransformData] wrapped script: %s", wrappedScript)
	}

	// 执行转换脚本
	log.Printf("[TransformData] executing wrapped script...")
	value, err := vm.RunString(wrappedScript)
	if err != nil {
		log.Printf("[TransformData] script execution failed: %v", err)
		log.Printf("[TransformData] rawData that caused error (first 500 chars): %s", rawData[:min(500, len(rawData))])
		return nil, fmt.Errorf("script execution failed: %v", err)
	}
	log.Printf("[TransformData] script executed successfully, value type: %T", value)

	// 尝试获取值的字符串表示
	if value != nil {
		if strValue := value.String(); strValue != "" {
			log.Printf("[TransformData] value as string (first 200 chars): %s", strValue[:min(200, len(strValue))])
		}
	}

	// 尝试将结果转换为Go的[]string
	var result []string

	// 首先尝试直接导出为[]string
	if err := vm.ExportTo(value, &result); err == nil && result != nil && len(result) > 0 {
		log.Printf("[TransformData] successfully exported as []string, count: %d", len(result))
		return result, nil
	} else {
		log.Printf("[TransformData] failed to export as []string: %v", err)
	}

	// 如果失败，尝试将结果转换为JSON字符串，然后解析
	jsonValue := vm.Get("JSON")
	if jsonValue != nil {
		stringify, ok := goja.AssertFunction(jsonValue.ToObject(vm).Get("stringify"))
		if ok {
			jsonStr, err := stringify(goja.Undefined(), value)
			if err == nil {
				log.Printf("[TransformData] JSON stringify result: %s", jsonStr.String())
				if err := json.Unmarshal([]byte(jsonStr.String()), &result); err == nil && len(result) > 0 {
					log.Printf("[TransformData] successfully parsed JSON string, count: %d", len(result))
					return result, nil
				} else {
					log.Printf("[TransformData] failed to unmarshal JSON string: %v", err)
				}
			} else {
				log.Printf("[TransformData] failed to stringify: %v", err)
			}
		}
	}

	// 最后尝试：如果value是数组，手动提取
	exported := value.Export()
	log.Printf("[TransformData] exported value type: %T, value: %v", exported, exported)
	if arr, ok := exported.([]interface{}); ok {
		log.Printf("[TransformData] value is []interface{}, length: %d", len(arr))
		result = make([]string, 0, len(arr))
		for i, item := range arr {
			if str, ok := item.(string); ok {
				result = append(result, str)
			} else {
				log.Printf("[TransformData] item[%d] is not string, type: %T, value: %v", i, item, item)
				result = append(result, fmt.Sprintf("%v", item))
			}
		}
		if len(result) > 0 {
			log.Printf("[TransformData] successfully extracted from []interface{}, count: %d", len(result))
			return result, nil
		}
	}

	log.Printf("[TransformData] all conversion attempts failed, value type: %T, value: %v", exported, exported)
	return nil, fmt.Errorf("script did not return a valid string array")
}

// TestTaskConfig 测试任务配置（执行curl和转换脚本）
func (s *ScheduledTaskService) TestTaskConfig(curlCommand, transformScript string) ([]string, error) {
	log.Printf("[TestTaskConfig] starting test, curlCommand length: %d, transformScript length: %d", len(curlCommand), len(transformScript))

	// 执行curl命令
	rawData, err := s.ExecuteCurlCommand(curlCommand)
	// 即使有错误，如果rawData不为空，也继续处理
	if err != nil {
		log.Printf("[TestTaskConfig] curl execution had error: %v, rawData length: %d", err, len(rawData))
		// 如果rawData为空，返回错误
		if len(rawData) == 0 {
			return nil, fmt.Errorf("curl execution failed: %v", err)
		}
		// 如果有数据，继续处理（curl可能只是有警告）
		log.Printf("[TestTaskConfig] continuing despite curl error, rawData is not empty")
	}

	// 如果rawData为空，返回错误
	if len(rawData) == 0 {
		log.Printf("[TestTaskConfig] curl execution returned empty data")
		return nil, fmt.Errorf("curl execution returned empty data")
	}

	log.Printf("[TestTaskConfig] curl execution successful, rawData length: %d, starting data transformation", len(rawData))

	// 执行转换脚本
	result, err := s.TransformData(rawData, transformScript)
	if err != nil {
		log.Printf("[TestTaskConfig] data transformation failed: %v", err)
		return nil, fmt.Errorf("data transformation failed: %v", err)
	}

	log.Printf("[TestTaskConfig] test successful, result count: %d", len(result))
	return result, nil
}

// GetTaskExecutions 获取任务执行历史
func (s *ScheduledTaskService) GetTaskExecutions(taskID uint, page, pageSize int) ([]model.TaskExecution, int64, error) {
	return s.executionRepo.ListByTaskID(taskID, page, pageSize)
}

// CreateTaskExecution 创建任务执行记录
func (s *ScheduledTaskService) CreateTaskExecution(execution *model.TaskExecution) error {
	return s.executionRepo.Create(execution)
}

// UpdateTaskExecution 更新任务执行记录
func (s *ScheduledTaskService) UpdateTaskExecution(execution *model.TaskExecution) error {
	return s.executionRepo.Update(execution)
}

// UpdateTaskRunTime 更新任务的执行时间
func (s *ScheduledTaskService) UpdateTaskRunTime(taskID uint, nextRunTime *time.Time) error {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return err
	}
	now := time.Now()
	task.LastRunAt = &now
	task.NextRunAt = nextRunTime
	return s.taskRepo.Update(task)
}
