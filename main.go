package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/sirupsen/logrus"
)

var (
	Config     ProxyConfig
	Log        *logrus.Logger
	configFile = flag.String("c", "./conf.yaml", "配置文件：conf.yaml")
)

func onExitSignal() {
	signalChan := make(chan os.Signal)
	// 监听系统服务退出信号
	signal.Notify(signalChan, syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGINT, os.Kill)
	for {
		signal := <-signalChan
		log.Println("Get Signal:%v\r\n", signal)
		switch signal {
		case syscall.SIGTERM, syscall.SIGINT, os.Kill:
			log.Fatal("系统退出。。。")
		}
	}
}
func main() {

	flag.Parse()
	
	// 解析配置
	parseConfigFile(*configFile)
    fmt.Println("parseConfig finish...")
	// 初始化日志模块
	initLogger()
    fmt.Println("Logger finish...")
	// 初始化代理的服务
	initBackendSvrs(Config.Backend)
    fmt.Println("Proxy finish...")
	// 系统退出信号监听
	go onExitSignal()

	// 初始化状态服务
	initStats()
	fmt.Printf("Listen server: %v \n", Config.Bind)
    fmt.Println("Start successful \n\n\n")
    
	// 初始化代理服务
	initProxy()
    
}
