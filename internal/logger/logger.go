package logger

import (
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

// Init 初始化日志记录器
func Init(logLevel string) {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Info 信息日志
func Info(v ...interface{}) {
	InfoLogger.Println(v...)
}

// Error 错误日志
func Error(v ...interface{}) {
	ErrorLogger.Println(v...)
}

// Debug 调试日志
func Debug(v ...interface{}) {
	DebugLogger.Println(v...)
}

// Infof 格式化信息日志
func Infof(format string, v ...interface{}) {
	InfoLogger.Printf(format, v...)
}

// Errorf 格式化错误日志
func Errorf(format string, v ...interface{}) {
	ErrorLogger.Printf(format, v...)
}

// Debugf 格式化调试日志
func Debugf(format string, v ...interface{}) {
	DebugLogger.Printf(format, v...)
}
