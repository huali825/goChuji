// zap_logger_test.go
package logger

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// TestZapLogger 测试 ZapLogger 的各个方法
func TestZapLogger(t *testing.T) {
	// 创建一个观察者来捕获日志输出
	observerCore, observedLogs := observer.New(zap.DebugLevel)
	zapLogger := zap.New(observerCore)

	// 创建 ZapLogger 实例
	logger := NewZapLogger(zapLogger)

	// 测试 Debug 方法
	logger.Debug("debug message", Field{Key: "key1", Val: "value1"})
	if observedLogs.Len() != 1 {
		t.Errorf("期望有1条日志，实际有%d条", observedLogs.Len())
	}

	debugLog := observedLogs.All()[0]
	if debugLog.Level != zap.DebugLevel {
		t.Errorf("期望日志级别为 DebugLevel，实际为 %v", debugLog.Level)
	}

	if debugLog.Message != "debug message" {
		t.Errorf("期望消息为 'debug message'，实际为 '%s'", debugLog.Message)
	}

	// 测试 Info 方法
	logger.Info("info message", Field{Key: "key2", Val: 123})
	if observedLogs.Len() != 2 {
		t.Errorf("期望有2条日志，实际有%d条", observedLogs.Len())
	}

	infoLog := observedLogs.All()[1]
	if infoLog.Level != zap.InfoLevel {
		t.Errorf("期望日志级别为 InfoLevel，实际为 %v", infoLog.Level)
	}

	if infoLog.Message != "info message" {
		t.Errorf("期望消息为 'info message'，实际为 '%s'", infoLog.Message)
	}

	// 测试 Warn 方法
	logger.Warn("warn message", Field{Key: "key3", Val: true})
	if observedLogs.Len() != 3 {
		t.Errorf("期望有3条日志，实际有%d条", observedLogs.Len())
	}

	warnLog := observedLogs.All()[2]
	if warnLog.Level != zap.WarnLevel {
		t.Errorf("期望日志级别为 WarnLevel，实际为 %v", warnLog.Level)
	}

	if warnLog.Message != "warn message" {
		t.Errorf("期望消息为 'warn message'，实际为 '%s'", warnLog.Message)
	}

	// 测试 Error 方法
	logger.Error("error message", Field{Key: "key4", Val: []string{"a", "b"}})
	if observedLogs.Len() != 4 {
		t.Errorf("期望有4条日志，实际有%d条", observedLogs.Len())
	}

	errorLog := observedLogs.All()[3]
	if errorLog.Level != zap.ErrorLevel {
		t.Errorf("期望日志级别为 ErrorLevel，实际为 %v", errorLog.Level)
	}

	if errorLog.Message != "error message" {
		t.Errorf("期望消息为 'error message'，实际为 '%s'", errorLog.Message)
	}
}

// TestZapLoggerToArgs 测试 toArgs 方法
func TestZapLoggerToArgs(t *testing.T) {
	observerCore, _ := observer.New(zap.DebugLevel)
	zapLogger := zap.New(observerCore)
	logger := NewZapLogger(zapLogger)

	// 准备测试数据
	fields := []Field{
		{Key: "string_key", Val: "string_value"},
		{Key: "int_key", Val: 42},
		{Key: "bool_key", Val: true},
		{Key: "slice_key", Val: []int{1, 2, 3}},
	}

	// 调用 toArgs 方法
	zapFields := logger.toArgs(fields)

	// 验证结果
	if len(zapFields) != len(fields) {
		t.Errorf("期望转换后的字段数量为 %d，实际为 %d", len(fields), len(zapFields))
	}
}

// TestZapLoggerWithMultipleFields 测试带有多个字段的日志记录
func TestZapLoggerWithMultipleFields(t *testing.T) {
	observerCore, observedLogs := observer.New(zap.DebugLevel)
	zapLogger := zap.New(observerCore)
	logger := NewZapLogger(zapLogger)

	// 记录带有多个字段的日志
	logger.Info("multiple fields test",
		Field{Key: "user_id", Val: 12345},
		Field{Key: "username", Val: "testuser"},
		Field{Key: "is_admin", Val: false})

	if observedLogs.Len() != 1 {
		t.Fatalf("期望有1条日志，实际有%d条", observedLogs.Len())
	}

	logEntry := observedLogs.All()[0]
	if logEntry.Message != "multiple fields test" {
		t.Errorf("期望消息为 'multiple fields test'，实际为 '%s'", logEntry.Message)
	}

	// 验证字段是否正确记录
	fields := logEntry.ContextMap()
	if len(fields) != 3 {
		t.Errorf("期望有3个字段，实际有%d个", len(fields))
	}

	if fields["user_id"] != int64(12345) {
		t.Errorf("期望 user_id 为 12345，实际为 %v", fields["user_id"])
	}

	if fields["username"] != "testuser" {
		t.Errorf("期望 username 为 'testuser'，实际为 '%v'", fields["username"])
	}

	if fields["is_admin"] != false {
		t.Errorf("期望 is_admin 为 false，实际为 %v", fields["is_admin"])
	}
}

// BenchmarkZapLogger 测试 ZapLogger 的性能
func BenchmarkZapLogger(b *testing.B) {
	// 创建一个无输出的 zap logger 用于性能测试
	zapLogger, _ := zap.NewDevelopment()
	logger := NewZapLogger(zapLogger)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("benchmark test", Field{Key: "count", Val: i})
	}
}
