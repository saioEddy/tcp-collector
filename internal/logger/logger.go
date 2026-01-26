package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
	"tcp-collector/internal/config"
)

// Init 初始化日志
func Init(cfg *config.LogConfig) error {
	var writers []io.Writer

	// 根据配置选择输出目标
	switch cfg.Output {
	case "console":
		writers = append(writers, os.Stdout)
	case "file":
		fileWriter, err := createFileWriter(cfg)
		if err != nil {
			return err
		}
		writers = append(writers, fileWriter)
	case "both":
		fileWriter, err := createFileWriter(cfg)
		if err != nil {
			return err
		}
		writers = append(writers, os.Stdout, fileWriter)
	default:
		writers = append(writers, os.Stdout)
	}

	// 设置多输出
	multiWriter := io.MultiWriter(writers...)
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags)

	return nil
}

// createFileWriter 创建文件写入器(带日志轮转)
func createFileWriter(cfg *config.LogConfig) (io.Writer, error) {
	// 确保日志目录存在
	logDir := filepath.Dir(cfg.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	// 使用lumberjack实现日志轮转
	return &lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    cfg.MaxSize, // MB
		MaxAge:     cfg.MaxDays, // days
		MaxBackups: 10,          // 保留最多10个备份文件
		LocalTime:  true,        // 使用本地时间
		Compress:   true,        // 压缩旧日志
	}, nil
}
