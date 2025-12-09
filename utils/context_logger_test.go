package utils

import (
	"errors"
	"testing"
	"time"
)

func TestTaskLogger_StructuredLogging(t *testing.T) {
	logger := NewTaskLogger("test-task")

	// Test StepStart
	logger.StepStart("step1")
	if len(logger.entries) != 1 {
		t.Errorf("Expected 1 entry after StepStart, got %d", len(logger.entries))
	}
	if len(logger.stepStack) != 1 {
		t.Errorf("Expected stepStack length 1, got %d", len(logger.stepStack))
	}

	// Test LogInfo
	logger.LogInfo("Test info message")
	if len(logger.entries) != 2 {
		t.Errorf("Expected 2 entries after LogInfo, got %d", len(logger.entries))
	}

	// Test LogProgress
	logger.LogProgress("/path/to/file", 1024, 2048, 50.0)
	if len(logger.entries) != 3 {
		t.Errorf("Expected 3 entries after LogProgress, got %d", len(logger.entries))
	}

	// Test StepSuccess
	logger.StepSuccess("step1")
	if len(logger.entries) != 4 {
		t.Errorf("Expected 4 entries after StepSuccess, got %d", len(logger.entries))
	}
	if len(logger.stepStack) != 0 {
		t.Errorf("Expected stepStack length 0 after StepSuccess, got %d", len(logger.stepStack))
	}

	// Test StepFailed
	logger.StepStart("step2")
	testErr := errors.New("test error")
	logger.StepFailed("step2", testErr)
	if len(logger.entries) != 6 {
		t.Errorf("Expected 6 entries after StepFailed, got %d", len(logger.entries))
	}
	if len(logger.stepStack) != 0 {
		t.Errorf("Expected stepStack length 0 after StepFailed, got %d", len(logger.stepStack))
	}

	// Test LogError
	logger.LogError(testErr, "Error occurred")
	if len(logger.entries) != 7 {
		t.Errorf("Expected 7 entries after LogError, got %d", len(logger.entries))
	}

	// Test GetEntries
	entries := logger.GetEntries()
	if len(entries) != 7 {
		t.Errorf("Expected GetEntries to return 7 entries, got %d", len(entries))
	}
}

func TestTaskLogger_EntryTypes(t *testing.T) {
	logger := NewTaskLogger("test-task")

	logger.StepStart("step1")
	logger.LogInfo("info")
	logger.LogProgress("/file", 100, 200, 50.0)
	logger.LogError(errors.New("error"), "error message")
	logger.StepSuccess("step1")

	entries := logger.GetEntries()

	expectedTypes := []LogEntryType{
		LogEntryTypeStep,
		LogEntryTypeInfo,
		LogEntryTypeProgress,
		LogEntryTypeError,
		LogEntryTypeStep,
	}

	for i, entry := range entries {
		if entry.Type != expectedTypes[i] {
			t.Errorf("Entry %d: expected type %s, got %s", i, expectedTypes[i], entry.Type)
		}
	}
}

func TestTaskLogger_StepStatus(t *testing.T) {
	logger := NewTaskLogger("test-task")

	logger.StepStart("step1")
	logger.StepSuccess("step1")
	logger.StepStart("step2")
	logger.StepFailed("step2", errors.New("failed"))

	entries := logger.GetEntries()

	if entries[0].StepStatus != StepStatusStart {
		t.Errorf("Expected first entry status to be 'start', got %s", entries[0].StepStatus)
	}
	if entries[1].StepStatus != StepStatusSuccess {
		t.Errorf("Expected second entry status to be 'success', got %s", entries[1].StepStatus)
	}
	if entries[2].StepStatus != StepStatusStart {
		t.Errorf("Expected third entry status to be 'start', got %s", entries[2].StepStatus)
	}
	if entries[3].StepStatus != StepStatusFailed {
		t.Errorf("Expected fourth entry status to be 'failed', got %s", entries[3].StepStatus)
	}
}

func TestTaskLogger_Timestamps(t *testing.T) {
	logger := NewTaskLogger("test-task")

	logger.LogInfo("message 1")
	time.Sleep(10 * time.Millisecond)
	logger.LogInfo("message 2")
	time.Sleep(10 * time.Millisecond)
	logger.LogInfo("message 3")

	entries := logger.GetEntries()

	// Verify timestamps are in order
	for i := 1; i < len(entries); i++ {
		if !entries[i].Timestamp.After(entries[i-1].Timestamp) {
			t.Errorf("Entry %d timestamp should be after entry %d", i, i-1)
		}
	}
}

func TestTaskLogger_BackwardCompatibility(t *testing.T) {
	logger := NewTaskLogger("test-task")

	// Test that old Log method still works
	logger.Log("Old style log message")

	entries := logger.GetEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry from Log method, got %d", len(entries))
	}
	if entries[0].Type != LogEntryTypeInfo {
		t.Errorf("Expected Log to create Info type entry, got %s", entries[0].Type)
	}
}

func TestTaskLogger_ProgressFields(t *testing.T) {
	logger := NewTaskLogger("test-task")

	filePath := "/test/file.txt"
	processed := int64(1024)
	total := int64(2048)
	percentage := 50.0

	logger.LogProgress(filePath, processed, total, percentage)

	entries := logger.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.FilePath != filePath {
		t.Errorf("Expected FilePath %s, got %s", filePath, entry.FilePath)
	}
	if entry.Processed != processed {
		t.Errorf("Expected Processed %d, got %d", processed, entry.Processed)
	}
	if entry.Total != total {
		t.Errorf("Expected Total %d, got %d", total, entry.Total)
	}
	if entry.Percentage != percentage {
		t.Errorf("Expected Percentage %.1f, got %.1f", percentage, entry.Percentage)
	}
}

func TestTaskLogger_ErrorFields(t *testing.T) {
	logger := NewTaskLogger("test-task")

	testErr := errors.New("test error")
	logger.LogError(testErr, "Error message")

	entries := logger.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Error != testErr {
		t.Errorf("Expected Error to be %v, got %v", testErr, entry.Error)
	}
	if entry.Type != LogEntryTypeError {
		t.Errorf("Expected Type to be error, got %s", entry.Type)
	}
}

// TestTaskLogger_LogStageSuccess 测试 LogStage 成功场景
func TestTaskLogger_LogStageSuccess(t *testing.T) {
	logger := NewTaskLogger("test-task")

	err := logger.LogStage("测试步骤", func() error {
		logger.LogInfo("步骤内部操作")
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	entries := logger.GetEntries()
	// 应该有 3 个条目：StepStart, LogInfo, StepSuccess
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// 验证第一个条目是 StepStart
	if entries[0].Type != LogEntryTypeStep || entries[0].StepStatus != StepStatusStart {
		t.Error("First entry should be StepStart")
	}

	// 验证第二个条目是 Info
	if entries[1].Type != LogEntryTypeInfo {
		t.Error("Second entry should be Info")
	}

	// 验证第三个条目是 StepSuccess
	if entries[2].Type != LogEntryTypeStep || entries[2].StepStatus != StepStatusSuccess {
		t.Error("Third entry should be StepSuccess")
	}
}

// TestTaskLogger_LogStageError 测试 LogStage 错误场景
func TestTaskLogger_LogStageError(t *testing.T) {
	logger := NewTaskLogger("test-task")

	testErr := errors.New("test error")
	err := logger.LogStage("测试步骤", func() error {
		logger.LogInfo("步骤内部操作")
		return testErr
	})

	if err != testErr {
		t.Errorf("Expected error %v, got %v", testErr, err)
	}

	entries := logger.GetEntries()
	// 应该有 3 个条目：StepStart, LogInfo, StepFailed
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// 验证最后一个条目是 StepFailed
	lastEntry := entries[len(entries)-1]
	if lastEntry.Type != LogEntryTypeStep || lastEntry.StepStatus != StepStatusFailed {
		t.Error("Last entry should be StepFailed")
	}

	if lastEntry.Error != testErr {
		t.Errorf("Expected error %v, got %v", testErr, lastEntry.Error)
	}
}

// TestTaskLogger_LogStagePanic 测试 LogStage 捕获 panic
func TestTaskLogger_LogStagePanic(t *testing.T) {
	logger := NewTaskLogger("test-task")

	err := logger.LogStage("测试步骤", func() error {
		logger.LogInfo("步骤内部操作")
		panic("test panic")
	})

	if err == nil {
		t.Error("Expected error from panic, got nil")
	}

	entries := logger.GetEntries()
	// 应该有 3 个条目：StepStart, LogInfo, StepFailed
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// 验证最后一个条目是 StepFailed
	lastEntry := entries[len(entries)-1]
	if lastEntry.Type != LogEntryTypeStep || lastEntry.StepStatus != StepStatusFailed {
		t.Error("Last entry should be StepFailed")
	}

	if lastEntry.Error == nil {
		t.Error("Expected error to be set from panic")
	}
}

// TestTaskLogger_LogStageNested 测试嵌套的 LogStage
func TestTaskLogger_LogStageNested(t *testing.T) {
	logger := NewTaskLogger("test-task")

	err := logger.LogStage("外层步骤", func() error {
		logger.LogInfo("外层操作")

		return logger.LogStage("内层步骤", func() error {
			logger.LogInfo("内层操作")
			return nil
		})
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	entries := logger.GetEntries()
	// 应该有 6 个条目：
	// 1. 外层 StepStart
	// 2. 外层 LogInfo
	// 3. 内层 StepStart
	// 4. 内层 LogInfo
	// 5. 内层 StepSuccess
	// 6. 外层 StepSuccess
	if len(entries) != 6 {
		t.Fatalf("Expected 6 entries, got %d", len(entries))
	}

	// 验证步骤顺序
	expectedTypes := []struct {
		entryType  LogEntryType
		stepStatus StepStatus
	}{
		{LogEntryTypeStep, StepStatusStart},
		{LogEntryTypeInfo, ""},
		{LogEntryTypeStep, StepStatusStart},
		{LogEntryTypeInfo, ""},
		{LogEntryTypeStep, StepStatusSuccess},
		{LogEntryTypeStep, StepStatusSuccess},
	}

	for i, expected := range expectedTypes {
		if entries[i].Type != expected.entryType {
			t.Errorf("Entry %d: expected type %s, got %s", i, expected.entryType, entries[i].Type)
		}
		if expected.stepStatus != "" && entries[i].StepStatus != expected.stepStatus {
			t.Errorf("Entry %d: expected status %s, got %s", i, expected.stepStatus, entries[i].StepStatus)
		}
	}
}
