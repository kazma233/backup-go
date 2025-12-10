package utils

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

// === Message ===

type MessageBody struct {
	Content string
	Date    time.Time
}

type Message struct {
	body []MessageBody
}

func (m *Message) String(sep string) string {
	if len(m.body) <= 0 {
		return ""
	}

	// 按时间戳排序消息，确保消息按正确顺序显示
	sort.Slice(m.body, func(i, j int) bool {
		return m.body[i].Date.Before(m.body[j].Date)
	})

	result := []string{}
	for _, msg := range m.body {
		result = append(result, fmt.Sprintf("%s: %s", msg.Date.Format("2006-01-02 15:04:05"), msg.Content))
	}

	return strings.Join(result, sep)
}

func (m *Message) Add(content string) {
	m.body = append(m.body, MessageBody{
		Content: content,
		Date:    time.Now(),
	})
}

func (m *Message) Clean() {
	m.body = make([]MessageBody, 0)
}

func (m *Message) Len() int {
	return len(m.body)
}

// === LogEntry ===

// LogEntryType 定义日志条目的类型
type LogEntryType string

const (
	LogEntryTypeStep     LogEntryType = "step"     // 步骤开始/结束
	LogEntryTypeProgress LogEntryType = "progress" // 进度信息
	LogEntryTypeInfo     LogEntryType = "info"     // 一般信息
	LogEntryTypeError    LogEntryType = "error"    // 错误信息
)

// StepStatus 定义步骤的状态
type StepStatus string

const (
	StepStatusStart   StepStatus = "start"
	StepStatusSuccess StepStatus = "success"
	StepStatusFailed  StepStatus = "failed"
)

// LogEntry 结构化的日志条目
type LogEntry struct {
	Type      LogEntryType
	Timestamp time.Time
	Message   string

	// 步骤相关字段
	StepName   string
	StepStatus StepStatus

	// 进度相关字段
	FilePath   string
	Processed  int64
	Total      int64
	Percentage float64

	// 错误相关字段
	Error error
}

// === TaskLogger ===

// TaskLogger 简化的任务日志记录器，只负责消息收集
type TaskLogger struct {
	taskID    string
	startTime time.Time
	message   Message
	entries   []LogEntry // 结构化日志条目
	stepStack []string   // 用于追踪嵌套步骤
}

// NewTaskLogger 创建新的任务日志记录器
func NewTaskLogger(taskID string) *TaskLogger {
	return &TaskLogger{
		taskID:    taskID,
		startTime: time.Now(),
		message: Message{
			body: make([]MessageBody, 0),
		},
		entries:   make([]LogEntry, 0),
		stepStack: make([]string, 0),
	}
}

// GetMessages 获取所有收集的消息
func (tl *TaskLogger) GetMessageAndClean() string {
	// 添加任务总结
	duration := time.Since(tl.startTime)
	tl.LogInfo("【%s】任务总耗时: %v", tl.taskID, duration)

	result := tl.message.String("\n")
	tl.message.Clean()
	return result
}

// === 结构化日志方法 ===

// LogInfo 记录一般信息
func (tl *TaskLogger) LogInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Type:      LogEntryTypeInfo,
		Timestamp: time.Now(),
		Message:   message,
	}
	tl.entries = append(tl.entries, entry)

	// 保持向后兼容
	tl.message.Add(message)
	log.Println(message)
}

// LogError 记录错误信息
func (tl *TaskLogger) LogError(err error, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Type:      LogEntryTypeError,
		Timestamp: time.Now(),
		Message:   message,
		Error:     err,
	}
	tl.entries = append(tl.entries, entry)

	// 保持向后兼容
	fullMessage := fmt.Sprintf("%s: %v", message, err)
	tl.message.Add(fullMessage)
	log.Println(fullMessage)
}

// LogProgress 记录进度信息
func (tl *TaskLogger) LogProgress(filePath string, processed, total int64, percentage float64) {
	entry := LogEntry{
		Type:       LogEntryTypeProgress,
		Timestamp:  time.Now(),
		FilePath:   filePath,
		Processed:  processed,
		Total:      total,
		Percentage: percentage,
		Message:    fmt.Sprintf("进度: %s (%.1f%%)", filePath, percentage),
	}
	tl.entries = append(tl.entries, entry)

	// 保持向后兼容
	message := fmt.Sprintf("进度: %s - %s / %s (%.1f%%)",
		filePath, FormatBytes(processed), FormatBytes(total), percentage)
	tl.message.Add(message)
	log.Println(message)
}

// GetEntries 返回所有日志条目
func (tl *TaskLogger) GetEntries() []LogEntry {
	return tl.entries
}

// GetStartTime 返回任务开始时间
func (tl *TaskLogger) GetStartTime() time.Time {
	return tl.startTime
}

// ExecuteStep 执行一个步骤，自动处理开始、成功和失败状态
// fn 可以返回 error 或者 panic，ExecuteStep 会自动捕获 panic 并转换为 error
func (tl *TaskLogger) ExecuteStep(stepName string, fn func() error) (err error) {
	tl.stepStart(stepName)

	defer func() {
		if r := recover(); r != nil {
			// 捕获 panic 并转换为 error
			err = fmt.Errorf("panic: %v", r)
			tl.stepFailed(stepName, err)
		} else if err != nil {
			// 函数返回了错误
			tl.stepFailed(stepName, err)
		} else {
			// 成功完成
			tl.stepSuccess(stepName)
		}
	}()

	return fn()
}

// stepStart 记录步骤开始并入栈
func (tl *TaskLogger) stepStart(stepName string) {
	tl.stepStack = append(tl.stepStack, stepName)
	entry := LogEntry{
		Type:       LogEntryTypeStep,
		Timestamp:  time.Now(),
		StepName:   stepName,
		StepStatus: StepStatusStart,
		Message:    fmt.Sprintf("开始: %s", stepName),
	}
	tl.entries = append(tl.entries, entry)

	// 保持向后兼容，同时记录到旧的消息系统
	message := fmt.Sprintf("【%s】%s 开始", tl.taskID, stepName)
	tl.message.Add(message)
	log.Println(message)
}

// stepSuccess 记录步骤成功并出栈
func (tl *TaskLogger) stepSuccess(stepName string) {
	// 出栈
	if len(tl.stepStack) > 0 {
		tl.stepStack = tl.stepStack[:len(tl.stepStack)-1]
	}

	entry := LogEntry{
		Type:       LogEntryTypeStep,
		Timestamp:  time.Now(),
		StepName:   stepName,
		StepStatus: StepStatusSuccess,
		Message:    fmt.Sprintf("完成: %s", stepName),
	}
	tl.entries = append(tl.entries, entry)

	// 保持向后兼容
	message := fmt.Sprintf("【%s】%s 完成", tl.taskID, stepName)
	tl.message.Add(message)
	log.Println(message)
}

// stepFailed 记录步骤失败并出栈
func (tl *TaskLogger) stepFailed(stepName string, err error) {
	// 出栈
	if len(tl.stepStack) > 0 {
		tl.stepStack = tl.stepStack[:len(tl.stepStack)-1]
	}

	entry := LogEntry{
		Type:       LogEntryTypeStep,
		Timestamp:  time.Now(),
		StepName:   stepName,
		StepStatus: StepStatusFailed,
		Message:    fmt.Sprintf("失败: %s", stepName),
		Error:      err,
	}
	tl.entries = append(tl.entries, entry)

	// 保持向后兼容
	message := fmt.Sprintf("【%s】%s 失败: %v", tl.taskID, stepName, err)
	tl.message.Add(message)
	log.Println(message)
}
