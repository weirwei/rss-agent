package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

func init() {
	// 确保日志目录存在
	if err := os.MkdirAll("logs", 0755); err != nil {
		panic(fmt.Sprintf("创建日志目录失败: %v", err))
	}

	// 打开日志文件
	logFile, err := os.OpenFile(
		filepath.Join("logs", "app.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		panic(fmt.Sprintf("打开日志文件失败: %v", err))
	}

	// 创建多重输出器，同时输出到文件和控制台
	infoMulti := io.MultiWriter(logFile, os.Stdout)
	errorMulti := io.MultiWriter(logFile, os.Stderr)

	// 创建日志记录器
	infoLogger = log.New(infoMulti, "INFO: ", log.Ldate|log.Ltime)
	errorLogger = log.New(errorMulti, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Info 记录信息日志
func Info(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}

// Error 记录错误日志
func Error(format string, v ...interface{}) {
	errorLogger.Output(2, fmt.Sprintf(format, v...))
}

// Fatal 记录致命错误并退出
func Fatal(format string, v ...interface{}) {
	errorLogger.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}
