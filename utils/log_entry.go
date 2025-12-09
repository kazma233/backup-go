package utils

import (
	"fmt"
	"time"
)

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

// FormatBytes 将字节数转换为人类可读的格式
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	if bytes < KB {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < MB {
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	} else if bytes < GB {
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	} else {
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	}
}

// FormatDuration 将时间间隔转换为易读格式
func FormatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d小时%d分%d秒", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%d分%d秒", minutes, seconds)
	} else {
		return fmt.Sprintf("%d秒", seconds)
	}
}

// FormatTimestamp 格式化时间戳
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatRelativeTime 格式化相对时间（用于步骤时间）
func FormatRelativeTime(start, current time.Time) string {
	elapsed := current.Sub(start)
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60
	seconds := int(elapsed.Seconds()) % 60

	return fmt.Sprintf("[%02d:%02d:%02d]", hours, minutes, seconds)
}
