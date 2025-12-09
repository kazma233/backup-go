package utils

import (
	"fmt"
	"strings"
	"time"
)

// MessageFormatter 定义消息格式化器接口
type MessageFormatter interface {
	Format(taskID string, startTime time.Time, entries []LogEntry) string
}

// PlainTextFormatter 纯文本格式化器，用于邮件和 Telegram
type PlainTextFormatter struct {
	showProgress bool // 是否显示详细进度
}

// NewPlainTextFormatter 创建新的纯文本格式化器
func NewPlainTextFormatter(showProgress bool) *PlainTextFormatter {
	return &PlainTextFormatter{
		showProgress: showProgress,
	}
}

// Format 将日志条目格式化为易读的纯文本消息
func (f *PlainTextFormatter) Format(taskID string, startTime time.Time, entries []LogEntry) string {
	var builder strings.Builder

	// 1. 生成标题部分
	f.formatHeader(&builder, taskID, entries)

	// 2. 生成摘要部分
	f.formatSummary(&builder, startTime, entries)

	// 3. 生成错误部分（如果有错误）
	if f.hasErrors(entries) {
		f.formatErrors(&builder, entries)
	}

	// 4. 生成详细日志部分
	f.formatDetails(&builder, startTime, entries)

	return builder.String()
}

// formatHeader 生成消息标题部分
func (f *PlainTextFormatter) formatHeader(builder *strings.Builder, taskID string, entries []LogEntry) {
	builder.WriteString("========================================\n")
	builder.WriteString(fmt.Sprintf("备份任务: %s\n", taskID))

	// 判断任务状态
	status := "✓ 成功"
	if f.hasErrors(entries) {
		status = "✗ 失败"
	}
	builder.WriteString(fmt.Sprintf("状态: %s\n", status))
	builder.WriteString("========================================\n\n")
}

// formatSummary 生成任务摘要部分
func (f *PlainTextFormatter) formatSummary(builder *strings.Builder, startTime time.Time, entries []LogEntry) {
	builder.WriteString("📊 任务摘要\n")
	builder.WriteString(fmt.Sprintf("  开始时间: %s\n", FormatTimestamp(startTime)))

	// 找到最后一个日志条目的时间作为结束时间
	var endTime time.Time
	if len(entries) > 0 {
		endTime = entries[len(entries)-1].Timestamp
	} else {
		endTime = startTime
	}

	builder.WriteString(fmt.Sprintf("  结束时间: %s\n", FormatTimestamp(endTime)))

	// 计算总耗时
	duration := endTime.Sub(startTime)
	builder.WriteString(fmt.Sprintf("  总耗时: %s\n", FormatDuration(duration)))

	// 提取关键指标（如压缩大小、上传方式等）
	f.extractKeyMetrics(builder, entries)

	builder.WriteString("\n")
}

// extractKeyMetrics 从日志条目中提取关键指标
func (f *PlainTextFormatter) extractKeyMetrics(builder *strings.Builder, entries []LogEntry) {
	// 查找压缩大小信息
	for _, entry := range entries {
		if entry.Type == LogEntryTypeInfo && strings.Contains(entry.Message, "压缩完成") {
			builder.WriteString(fmt.Sprintf("  %s\n", entry.Message))
		}
		if entry.Type == LogEntryTypeInfo && strings.Contains(entry.Message, "bucket") {
			builder.WriteString(fmt.Sprintf("  %s\n", entry.Message))
		}
	}
}

// formatErrors 生成错误部分
func (f *PlainTextFormatter) formatErrors(builder *strings.Builder, entries []LogEntry) {
	builder.WriteString("❌ 错误信息\n")

	for _, entry := range entries {
		if entry.Type == LogEntryTypeError || (entry.Type == LogEntryTypeStep && entry.StepStatus == StepStatusFailed) {
			if entry.StepName != "" {
				builder.WriteString(fmt.Sprintf("  步骤: %s\n", entry.StepName))
			}
			if entry.Error != nil {
				builder.WriteString(fmt.Sprintf("  错误: %v\n", entry.Error))
			}
			if entry.Message != "" {
				builder.WriteString(fmt.Sprintf("  详情: %s\n", entry.Message))
			}
			builder.WriteString(fmt.Sprintf("  时间: %s\n", FormatTimestamp(entry.Timestamp)))
			builder.WriteString("\n")
		}
	}
}

// formatDetails 生成详细日志部分
func (f *PlainTextFormatter) formatDetails(builder *strings.Builder, startTime time.Time, entries []LogEntry) {
	builder.WriteString("📝 执行详情\n\n")

	// 追踪当前步骤深度，用于缩进
	stepDepth := 0
	lastProgressFile := "" // 用于去重进度信息

	for _, entry := range entries {
		relativeTime := FormatRelativeTime(startTime, entry.Timestamp)

		switch entry.Type {
		case LogEntryTypeStep:
			f.formatStepEntry(builder, entry, relativeTime, &stepDepth)

		case LogEntryTypeProgress:
			// 只在 showProgress 为 true 或文件变化时显示进度
			if f.showProgress || entry.FilePath != lastProgressFile {
				f.formatProgressEntry(builder, entry, stepDepth)
				lastProgressFile = entry.FilePath
			}

		case LogEntryTypeInfo:
			f.formatInfoEntry(builder, entry, stepDepth)

		case LogEntryTypeError:
			f.formatErrorEntry(builder, entry, stepDepth)
		}
	}
}

// formatStepEntry 格式化步骤日志条目
func (f *PlainTextFormatter) formatStepEntry(builder *strings.Builder, entry LogEntry, relativeTime string, stepDepth *int) {
	indent := strings.Repeat("  ", *stepDepth)

	switch entry.StepStatus {
	case StepStatusStart:
		builder.WriteString(fmt.Sprintf("%s ▶ %s\n", relativeTime, entry.StepName))
		*stepDepth++

	case StepStatusSuccess:
		*stepDepth--
		if *stepDepth < 0 {
			*stepDepth = 0
		}
		indent = strings.Repeat("  ", *stepDepth)
		builder.WriteString(fmt.Sprintf("%s  ✓ %s\n", indent, entry.Message))
		if entry.Message == "" {
			builder.WriteString(fmt.Sprintf("%s  ✓ %s完成\n", indent, entry.StepName))
		}
		builder.WriteString("\n")

	case StepStatusFailed:
		*stepDepth--
		if *stepDepth < 0 {
			*stepDepth = 0
		}
		indent = strings.Repeat("  ", *stepDepth)
		builder.WriteString(fmt.Sprintf("%s  ✗ %s失败\n", indent, entry.StepName))
		if entry.Error != nil {
			builder.WriteString(fmt.Sprintf("%s    错误: %v\n", indent, entry.Error))
		}
		builder.WriteString("\n")
	}
}

// formatProgressEntry 格式化进度日志条目
func (f *PlainTextFormatter) formatProgressEntry(builder *strings.Builder, entry LogEntry, stepDepth int) {
	indent := strings.Repeat("  ", stepDepth)
	builder.WriteString(fmt.Sprintf("%s  进度: %s (%.1f%%)\n",
		indent, entry.FilePath, entry.Percentage))
}

// formatInfoEntry 格式化信息日志条目
func (f *PlainTextFormatter) formatInfoEntry(builder *strings.Builder, entry LogEntry, stepDepth int) {
	indent := strings.Repeat("  ", stepDepth)
	builder.WriteString(fmt.Sprintf("%s  %s\n", indent, entry.Message))
}

// formatErrorEntry 格式化错误日志条目
func (f *PlainTextFormatter) formatErrorEntry(builder *strings.Builder, entry LogEntry, stepDepth int) {
	indent := strings.Repeat("  ", stepDepth)
	builder.WriteString(fmt.Sprintf("%s  ❌ %s\n", indent, entry.Message))
	if entry.Error != nil {
		builder.WriteString(fmt.Sprintf("%s     错误: %v\n", indent, entry.Error))
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
