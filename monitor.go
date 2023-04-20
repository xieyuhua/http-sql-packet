package main

import (
	"fmt"
	"net/http"
// 	"log"
)

// 查询监控信息的接口
func statsHandler(w http.ResponseWriter, r *http.Request) {
	_str := ""
	_str += fmt.Sprintf("Server max %d, WaitQueueLen %d, connecting num:%d \n\n", Config.MaxConn, Config.WaitQueueLen, len(Arrip))
	
	for _, vv := range Arrip {
		_str += fmt.Sprintf("connecting ip :%s \n", vv)
	}
	
	for _, v := range BackendSvrs {
		_str += fmt.Sprintf("\nServer:%s FailTimes:%d isUp:%t\n", v.identify, v.failTimes, v.isLive)
	}
	
// 	log.Println(Arrip)
	
// 	fmt.Fprintf(w, "%s", _str)
	w.Write([]byte(_str))
}

// 初始化监控服务地址
func initStats() {
	Log.Infof("Start monitor on addr %s", Config.Stats)

	go func() {
		http.HandleFunc("/", statsHandler)
		http.ListenAndServe(Config.Stats, nil)
	}()
}
