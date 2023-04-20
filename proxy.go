package main

import (
	"net"
	"time"
	"fmt"
	"sync"
)

var Arrip map[string]string
var countGuard sync.Mutex

// 初始化代理服务
func initProxy() {
	Log.Infof("Proxying %s -> %s\n", Config.Bind, Config.Backend)
	server, err := net.Listen("tcp", Config.Bind)
	if err != nil {
		Log.Fatal(err)
	}
	Arrip = make(map[string]string)
	// 等待的队列长度
	waitQueue := make(chan net.Conn, Config.WaitQueueLen)
	// 最大并发连接
	connPools := make(chan bool, Config.MaxConn)
	for i := 0; i < Config.MaxConn; i++ {
		connPools <- true
	}
	// 等待连接处理
	go waitConn(waitQueue, connPools)
	// 接收连接并抛给管道处理
	
    str := "Connection refused"
    buf := make([]byte, len(str)+1)
    copy(buf, str)
    buf[len(str)] = '\n'
	
	for {
		conn, err := server.Accept()
		if err != nil {
			Log.Error(err)
			continue
		}
		Log.Infof("Received connection from %s.\n", conn.RemoteAddr())
		
		if (len(Arrip)+1)>(Config.WaitQueueLen + Config.MaxConn) {
    		conn.Write(buf)
    		conn.Close()
		}else{
    		fromRemoteAddr := fmt.Sprintf("%s", conn.RemoteAddr())
    		countGuard.Lock()
    		Arrip[fromRemoteAddr] = fromRemoteAddr
    		countGuard.Unlock()
    		
		    waitQueue <- conn
		}

	}
}

// 连接数控制
func waitConn(waitQueue chan net.Conn, connPools chan bool) {
	for conn := range waitQueue {
		// 接收一个链接，连接池释放一个
		<-connPools
		go func(conn net.Conn) {
			handleConn(conn)
			// 链接处理完毕，增加
			connPools <- true
			fromRemoteAddr := fmt.Sprintf("%s", conn.RemoteAddr())
			countGuard.Lock()
			delete(Arrip, fromRemoteAddr)
			countGuard.Unlock()
			Log.Infof("Closed connection from %s.\n", conn.RemoteAddr())
		}(conn)
	}
}

// 处理连接
func handleConn(conn net.Conn) {
	defer conn.Close()
	// 根据链接哈希选择机器
	proxySvr, ok := getBackendSvr(conn)
	if !ok {
		return
	}
	// 链接远程代理服务器
	remote, err := net.Dial("tcp", proxySvr.identify)
	if err != nil {
		Log.Error(err)
		proxySvr.failTimes++
		return
	}
	// 等待双向连接完成
	complete := make(chan bool, 2)
	oneSwitch := make(chan bool, 1)
	otherSwitch := make(chan bool, 1)
	// 将当前客户端链接发送的数据发送给远程被代理的服务器
	Log.Infof("from %s to %s.\n", conn.RemoteAddr(), remote.RemoteAddr())
	go transaction(conn, remote, complete, oneSwitch, otherSwitch, true)
	// 将远程服务返回的数据返回给客户端
	go transaction(remote, conn, complete, otherSwitch, oneSwitch, false)
	<-complete
	<-complete
	remote.Close()
}

// 数据交换传输（从from读数据，再写入to）
func transaction(from, to net.Conn, complete, oneSwitch, otherSwitch chan bool, out bool) {
	var err error
	var read int
	bytes := make([]byte, 5120)
	for {
		select {
		case <-otherSwitch:
			complete <- true
			return
		default:
			timeOutSec := time.Duration(Config.Timeout) * time.Second
			// 设置超时时间
			from.SetReadDeadline(time.Now().Add(timeOutSec))
			read, err = from.Read(bytes)
			if err != nil {
				complete <- true
				oneSwitch <- true
				return
			}
			//客户端请求
			if out {
			    fromRemoteAddr := fmt.Sprintf("client:%s", from.RemoteAddr())
			    toRemoteAddr   := fmt.Sprintf("Server:%s", to.RemoteAddr())
			    if Config.Type=="mysql" {
			        proxyLog(fromRemoteAddr, toRemoteAddr,read, bytes)
			    }
			    if Config.Type=="redis" {
			        fmt.Println(string(bytes))
			    }
			    
			}
			// 设置超时时间
			to.SetWriteDeadline(time.Now().Add(timeOutSec))
			_, err = to.Write(bytes[:read])
			if err != nil {
				complete <- true
				oneSwitch <- true
				return
			}
		}
	}
}
