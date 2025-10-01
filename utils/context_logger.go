package utils

import (
	"fmt"
	"log"
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

	result := []string{}
	for _, m := range m.body {
		result = append(result, fmt.Sprintf("%s: %s", m.Date.Format("2006-01-02 15:04:05"), m.Content))
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

// === TaskLogger ===

// TaskLogger 简化的任务日志记录器，只负责消息收集
type TaskLogger struct {
	taskID    string
	startTime time.Time
	message   Message
}

// NewTaskLogger 创建新的任务日志记录器
func NewTaskLogger(taskID string) *TaskLogger {
	return &TaskLogger{
		taskID:    taskID,
		startTime: time.Now(),
		message: Message{
			body: make([]MessageBody, 0),
		},
	}
}

// Log 记录消息（唯一需要的日志方法）
func (tl *TaskLogger) Log(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	tl.message.Add(message)
	log.Println(message)
}

// Execute 执行函数并自动处理错误
func (tl *TaskLogger) Execute(stepName string, fn func() error) error {
	tl.Log("【%s】%s 开始", tl.taskID, stepName)

	if err := fn(); err != nil {
		tl.Log("【%s】%s 失败: %v", tl.taskID, stepName, err)
		return err
	}

	tl.Log("【%s】%s 完成", tl.taskID, stepName)
	return nil
}

// ExecuteWithPanic 执行函数并自动处理 panic
func (tl *TaskLogger) ExecuteWithPanic(stepName string, fn func()) {
	tl.Log("【%s】%s 开始", tl.taskID, stepName)

	defer func() {
		if r := recover(); r != nil {
			tl.Log("【%s】%s 异常: %v", tl.taskID, stepName, r)
		} else {
			tl.Log("【%s】%s 完成", tl.taskID, stepName)
		}
	}()

	fn()
}

// GetMessages 获取所有收集的消息
func (tl *TaskLogger) GetMessageAndClean() string {
	// 添加任务总结
	duration := time.Since(tl.startTime)
	tl.Log("【%s】任务总耗时: %v", tl.taskID, duration)

	result := tl.message.String("\n")
	tl.message.Clean()
	return result
}
