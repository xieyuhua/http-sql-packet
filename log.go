package main

import (
	"os"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/sirupsen/logrus"
)

// 初始化日志模块
func initLogger() error {
	logFilePath := Config.Log.Path
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	// 解析日志记录的等级信息
	level, err := logrus.ParseLevel(Config.Log.Level)
	if err != nil {
		return err
	}
	// 初始化日志结构
	Log = &logrus.Logger{
		Out:       file,
		Level:     level,
		Formatter: new(logrus.JSONFormatter),
	}
	
	
    //日志
    Logs.SetFormatter(&logrus.JSONFormatter{})
    
	logger := &lumberjack.Logger{
		Filename:   "logs/logrus.log",
		MaxSize:    50,  // 日志文件大小，单位是 MB
		MaxBackups: 3,    // 最大过期日志保留个数
		MaxAge:     30,   // 保留过期文件最大时间，单位 天
		Compress:   true, // 是否压缩日志，默认是不压缩。这里设置为true，压缩日志
	}
	Logs.SetOutput(logger) // logrus 设置日志的输出方式
	
	
	return nil
}
