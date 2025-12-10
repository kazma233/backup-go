package utils

import (
	"fmt"
	"strings"
	"time"
)

// === MessageFormatter ===

// MessageFormatter 定义消息格式化器接口
type MessageFormatter interface {
	Format(taskID string, startTime time.Time, entries []LogEntry) string
}

// PlainTextFormatter 纯文本格式化器，用于邮件和 Telegram
type PlainTextFormatter struct {
	showProgress bool // 是否显示详细进度
	builder      strings.Builder
}

// NewPlainTextFormatter 创建新的纯文本格式化器
func NewPlainTextFormatter(showProgress bool) *PlainTextFormatter {
	return &PlainTextFormatter{
		showProgress: showProgress,
	}
}

// Format 将日志条目格式化为易读的纯文本消息
func (f *PlainTextFormatter) Format(taskID string, startTime time.Time, entries []LogEntry) string {
	// 重置 builder
	f.builder.Reset()

	// 1. 生成标题部分
	f.formatHeader(taskID, entries)

	// 2. 生成摘要部分
	f.formatSummary(startTime, entries)

	// 3. 生成错误部分（如果有错误）
	if f.hasErrors(entries) {
		f.formatErrors(entries)
	}

	// 4. 生成详细日志部分
	f.formatDetails(startTime, entries)

	return f.builder.String()
}

// formatHeader 生成消息标题部分
func (f *PlainTextFormatter) formatHeader(taskID string, entries []LogEntry) {
	f.builder.WriteString("========================================\n")
	fmt.Fprintf(&f.builder, "备份任务: %s\n", taskID)

	// 判断任务状态
	status := "✓ 成功"
	if f.hasErrors(entries) {
		status = "✗ 失败"
	}
	fmt.Fprintf(&f.builder, "状态: %s\n", status)
	f.builder.WriteString("========================================\n\n")
}

// formatSummary 生成任务摘要部分
func (f *PlainTextFormatter) formatSummary(startTime time.Time, entries []LogEntry) {
	f.builder.WriteString("📊 任务摘要\n")
	fmt.Fprintf(&f.builder, "  开始时间: %s\n", FormatTimestamp(startTime))

	// 找到最后一个日志条目的时间作为结束时间
	var endTime time.Time
	if len(entries) > 0 {
		endTime = entries[len(entries)-1].Timestamp
	} else {
		endTime = startTime
	}

	fmt.Fprintf(&f.builder, "  结束时间: %s\n", FormatTimestamp(endTime))

	// 计算总耗时
	duration := endTime.Sub(startTime)
	fmt.Fprintf(&f.builder, "  总耗时: %s\n", FormatDuration(duration))

	// 提取关键指标（如压缩大小、上传方式等）
	f.extractKeyMetrics(entries)

	f.builder.WriteString("\n")
}

// extractKeyMetrics 从日志条目中提取关键指标
func (f *PlainTextFormatter) extractKeyMetrics(entries []LogEntry) {
	// 查找压缩大小信息
	for _, entry := range entries {
		if entry.Type == LogEntryTypeInfo && strings.Contains(entry.Message, "压缩完成") {
			fmt.Fprintf(&f.builder, "  %s\n", entry.Message)
		}
		if entry.Type == LogEntryTypeInfo && strings.Contains(entry.Message, "bucket") {
			fmt.Fprintf(&f.builder, "  %s\n", entry.Message)
		}
	}
}

// formatErrors 生成错误部分
func (f *PlainTextFormatter) formatErrors(entries []LogEntry) {
	f.builder.WriteString("❌ 错误信息\n")

	for _, entry := range entries {
		if entry.Type == LogEntryTypeError || (entry.Type == LogEntryTypeStep && entry.StepStatus == StepStatusFailed) {
			if entry.StepName != "" {
				fmt.Fprintf(&f.builder, "  步骤: %s\n", entry.StepName)
			}
			if entry.Error != nil {
				fmt.Fprintf(&f.builder, "  错误: %v\n", entry.Error)
			}
			if entry.Message != "" {
				fmt.Fprintf(&f.builder, "  详情: %s\n", entry.Message)
			}
			fmt.Fprintf(&f.builder, "  时间: %s\n", FormatTimestamp(entry.Timestamp))
			f.builder.WriteString("\n")
		}
	}
}

// formatDetails 生成详细日志部分
func (f *PlainTextFormatter) formatDetails(startTime time.Time, entries []LogEntry) {
	f.builder.WriteString("📝 执行详情\n\n")

	// 追踪当前步骤深度，用于缩进
	stepDepth := 0
	lastProgressFile := "" // 用于去重进度信息

	for _, entry := range entries {
		relativeTime := FormatRelativeTime(startTime, entry.Timestamp)

		switch entry.Type {
		case LogEntryTypeStep:
			f.formatStepEntry(entry, relativeTime, &stepDepth)

		case LogEntryTypeProgress:
			// 只在 showProgress 为 true 或文件变化时显示进度
			if f.showProgress || entry.FilePath != lastProgressFile {
				f.formatProgressEntry(entry, stepDepth)
				lastProgressFile = entry.FilePath
			}

		case LogEntryTypeInfo:
			f.formatInfoEntry(entry, stepDepth)

		case LogEntryTypeError:
			f.formatErrorEntry(entry, stepDepth)
		}
	}
}

// formatStepEntry 格式化步骤日志条目
func (f *PlainTextFormatter) formatStepEntry(entry LogEntry, relativeTime string, stepDepth *int) {
	switch entry.StepStatus {
	case StepStatusStart:
		indent := strings.Repeat("  ", *stepDepth)
		fmt.Fprintf(&f.builder, "%s%s ▶ %s\n", indent, relativeTime, entry.StepName)
		*stepDepth++

	case StepStatusSuccess:
		*stepDepth--
		if *stepDepth < 0 {
			*stepDepth = 0
		}
		indent := strings.Repeat("  ", *stepDepth)
		fmt.Fprintf(&f.builder, "%s  ✓ %s\n", indent, entry.Message)
		if entry.Message == "" {
			fmt.Fprintf(&f.builder, "%s  ✓ %s完成\n", indent, entry.StepName)
		}
		f.builder.WriteString("\n")

	case StepStatusFailed:
		*stepDepth--
		if *stepDepth < 0 {
			*stepDepth = 0
		}
		indent := strings.Repeat("  ", *stepDepth)
		fmt.Fprintf(&f.builder, "%s  ✗ %s失败\n", indent, entry.StepName)
		if entry.Error != nil {
			fmt.Fprintf(&f.builder, "%s    错误: %v\n", indent, entry.Error)
		}
		f.builder.WriteString("\n")
	}
}

// formatProgressEntry 格式化进度日志条目
func (f *PlainTextFormatter) formatProgressEntry(entry LogEntry, stepDepth int) {
	indent := strings.Repeat("  ", stepDepth)
	fmt.Fprintf(&f.builder, "%s  进度: %s (%.1f%%)\n", indent, entry.FilePath, entry.Percentage)
}

// formatInfoEntry 格式化信息日志条目
func (f *PlainTextFormatter) formatInfoEntry(entry LogEntry, stepDepth int) {
	indent := strings.Repeat("  ", stepDepth)
	fmt.Fprintf(&f.builder, "%s  %s\n", indent, entry.Message)
}

// formatErrorEntry 格式化错误日志条目
func (f *PlainTextFormatter) formatErrorEntry(entry LogEntry, stepDepth int) {
	indent := strings.Repeat("  ", stepDepth)
	fmt.Fprintf(&f.builder, "%s  ❌ %s\n", indent, entry.Message)
	if entry.Error != nil {
		fmt.Fprintf(&f.builder, "%s     错误: %v\n", indent, entry.Error)
	}
}

// hasErrors 检查日志条目中是否有错误
func (f *PlainTextFormatter) hasErrors(entries []LogEntry) bool {
	for _, entry := range entries {
		if entry.Type == LogEntryTypeError || (entry.Type == LogEntryTypeStep && entry.StepStatus == StepStatusFailed) {
			return true
		}
	}
	return false
}

// === tools ===

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
