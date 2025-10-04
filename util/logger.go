package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

const (
	// LevelError 错误
	LevelError = iota
	// LevelWarning 警告
	LevelWarning
	// LevelInformational 提示
	LevelInformational
	// LevelDebug 除错
	LevelDebug
)

var logger *Logger

// Location 全局时区设置
var Location = time.Local

// Logger 日志
type Logger struct {
	level int
	out   io.Writer
}

// Println 打印
func (ll *Logger) Println(level string, msg string) {
	// 指向调用 Error/Info 的上层调用处
	_, file, line, _ := runtime.Caller(2)
	fileName := filepath.Base(file)
	if ll.out == nil {
		ll.out = os.Stdout
	}
	fmt.Fprintf(ll.out, "[%s] %s | %s:%d | %s \n", level, GetCurrentTime().Format("2006-01-02 15:04:05"), fileName, line, msg)
}

// Panic 极端错误
func (ll *Logger) Panic(format string, v ...interface{}) {
	if LevelError > ll.level {
		return
	}
	msg := fmt.Sprintf(format, v...)
	ll.Println("Panic", msg)
	os.Exit(0)
}

// Error 错误
func (ll *Logger) Error(format string, v ...interface{}) {
	if LevelError > ll.level {
		return
	}
	msg := fmt.Sprintf(format, v...)
	ll.Println("Error", msg)
}

// Warning 警告
func (ll *Logger) Warning(format string, v ...interface{}) {
	if LevelWarning > ll.level {
		return
	}
	msg := fmt.Sprintf(format, v...)
	ll.Println("Warning", msg)
}

// Info 信息
func (ll *Logger) Info(format string, v ...interface{}) {
	if LevelInformational > ll.level {
		return
	}
	msg := fmt.Sprintf(format, v...)
	ll.Println("Info", msg)
}

// Debug 校验
func (ll *Logger) Debug(format string, v ...interface{}) {
	if LevelDebug > ll.level {
		return
	}
	msg := fmt.Sprintf(format, v...)
	ll.Println("Debug", msg)
}

// BuildLogger 构建logger
func BuildLogger(level string) {
	intLevel := LevelError
	switch level {
	case "error":
		intLevel = LevelError
	case "warning":
		intLevel = LevelWarning
	case "info":
		intLevel = LevelInformational
	case "debug":
		intLevel = LevelDebug
	}
	// 读取文件日志配置（可选）
	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		// 默认写入到 ./logs/app.log（若未指定）
		cwd, _ := os.Getwd()
		logFile = filepath.Join(cwd, "logs", "app.log")
	}

	// 确保目录存在
	if dir := filepath.Dir(logFile); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	// 可配置项
	maxSizeMB := parseIntWithDefault(os.Getenv("LOG_MAX_SIZE_MB"), 100)  // 单个文件最大尺寸（MB）
	maxBackups := parseIntWithDefault(os.Getenv("LOG_MAX_BACKUPS"), 7)   // 备份数量
	maxAgeDays := parseIntWithDefault(os.Getenv("LOG_MAX_AGE_DAYS"), 30) // 保留天数
	compress := parseBoolWithDefault(os.Getenv("LOG_COMPRESS"), true)    // 是否压缩历史日志
	alsoConsole := parseBoolWithDefault(os.Getenv("LOG_CONSOLE"), true)  // 是否同时输出到控制台

	fileWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   compress,
	}

	var out io.Writer = fileWriter
	if alsoConsole {
		out = io.MultiWriter(os.Stdout, fileWriter)
	}

	l := Logger{
		level: intLevel,
		out:   out,
	}
	logger = &l
}

// Log 返回日志对象
func Log() *Logger {
	if logger == nil {
		l := Logger{
			level: LevelDebug,
			out:   os.Stdout,
		}
		logger = &l
	}
	return logger
}

// LogWriter 返回当前日志输出 writer（用于复用到第三方库，如 GORM）
func LogWriter() io.Writer {
	l := Log()
	if l.out == nil {
		l.out = os.Stdout
	}
	return l.out
}

// parseIntWithDefault 将字符串解析为 int，失败则返回默认值
func parseIntWithDefault(raw string, def int) int {
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}

// parseBoolWithDefault 将字符串解析为 bool，支持 1/0/true/false/yes/no，失败返回默认值
func parseBoolWithDefault(raw string, def bool) bool {
	if raw == "" {
		return def
	}
	lower := strings.ToLower(strings.TrimSpace(raw))
	switch lower {
	case "1", "true", "t", "yes", "y":
		return true
	case "0", "false", "f", "no", "n":
		return false
	default:
		return def
	}
}

// GetCurrentTime 获取当前时间
func GetCurrentTime() time.Time {
	return time.Now().In(Location)
}
