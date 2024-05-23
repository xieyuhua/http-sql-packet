package main

import (
	"net"
	"time"
	"fmt"
	"sync"
	"strings"
)

var Onlinenum int
// var List map[string]*IPStruct

type IPStruct struct {
	Time int64    
	Title string  
}
var TotalBytesLock sync.Mutex

// 初始化代理服务
func initProxy() {
	Log.Infof("Proxying %s -> %s \n", Config.Bind, Config.Backend)
	server, err := net.Listen("tcp", Config.Bind)
	if err != nil {
		Log.Fatal(err)
	}
	
	
	// 等待的队列长度
	waitQueue := make(chan net.Conn, Config.WaitQueueLen)
	// 最大并发连接
	connPools := make(chan bool, Config.MaxConn)
	for i := 0; i < Config.MaxConn; i++ {
		connPools <- true
	}
	
	
	Onlinenum = 0
	// 等待连接处理
	go waitConn(waitQueue, connPools)
	// 接收连接并抛给管道处理
	for {
		conn, err := server.Accept()
		if err != nil {
			Log.Error(err)
			continue
		}
		Log.Infof("Received connection from %s.\n", conn.RemoteAddr())
		waitQueue <- conn
	}
}

// 连接数控制
func waitConn(waitQueue chan net.Conn, connPools chan bool) {
	for conn := range waitQueue {
		// 接收一个链接，连接池释放一个
		<-connPools
		Onlinenum = Onlinenum + 1
		go func(conn net.Conn) {
			handleConn(conn)
			// 链接处理完毕，增加
			connPools <- true
			Onlinenum = Onlinenum - 1
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
	
	List :=  make(map[string]*IPStruct)
	
	go transaction(conn, remote, complete, oneSwitch, otherSwitch, true, List)
	// 将远程服务返回的数据返回给客户端
	go transaction(remote, conn, complete, otherSwitch, oneSwitch, false, List)
	<-complete
	<-complete
	
    TotalBytesLock.Lock()
    delete(List,  fmt.Sprintf("%v:%v", conn.RemoteAddr(), remote.RemoteAddr()))
    delete(List,  fmt.Sprintf("%v:%v", remote.RemoteAddr(), conn.RemoteAddr()))
    TotalBytesLock.Unlock()
    
	remote.Close()
}


// 数据交换传输（从from读数据，再写入to）
func transaction(from, to net.Conn, complete, oneSwitch, otherSwitch chan bool, out bool, List map[string]*IPStruct) {
	var err error
	var read int
	buffer := make([]byte, Config.BulkSize)
	for {
		select {
		case <-otherSwitch:
			complete <- true
			return
		default:
			timeOutSec := time.Duration(Config.Timeout) * time.Second
			// 设置超时时间
			from.SetReadDeadline(time.Now().Add(timeOutSec))
			read, err = from.Read(buffer)
			if err != nil {
				complete <- true
				oneSwitch <- true
				return
			}
			//客户端请求
			if out {
			    
			    str := ""
			 //   fmt.Println(string(buffer))
			    fromRemoteAddr := fmt.Sprintf("client:%s", from.RemoteAddr())
			    toRemoteAddr   := fmt.Sprintf("Server:%s", to.RemoteAddr())
			    if Config.Type=="mysql" {
			       str = proxyLog(fromRemoteAddr, toRemoteAddr, read, buffer)
			    }
			    
			    if Config.Type=="redis" {
			        str = string(buffer)
			    }
			    if Config.Type=="oracle" {
			       str = ParseOracleSQL(fromRemoteAddr, toRemoteAddr, read, buffer)
			    }
			    //自助收银机参数更换处理
			    if Config.Type=="http" {
			       str = string(buffer)
			       buffer = []byte( strings.Replace(string(buffer), `busno":"2040"`, `"busno":"9013"`, -1) )
			       buffer = []byte( strings.Replace(string(buffer), `shopno":"2040"`, `"shopno":"9013"`, -1) )
			     //  fmt.Println(string(buffer))
			    }
			 //   fmt.Println(string(buffer))
			    //记录时间
			 //   if strings.TrimSpace(str) != "" {
			 	TotalBytesLock.Lock()

			        List[fmt.Sprintf("%v:%v", from.RemoteAddr(), to.RemoteAddr())] =  &IPStruct{Time:time.Now().UnixNano(), Title:str} 
			 //   }
                TotalBytesLock.Unlock()
			}else{
			    //计算时间
			    TotalBytesLock.Lock()
			    start, ok :=  List[ fmt.Sprintf("%v:%v", to.RemoteAddr(), from.RemoteAddr() ) ]
			    TotalBytesLock.Unlock()
			    if ok && strings.TrimSpace(start.Title) != "" {  
                    elapsed := float64( time.Now().UnixNano() - start.Time )/ float64(time.Millisecond)
                    
                    //记录耗时比较长的数据
                    if int(elapsed) > int(Config.SlowTime) {
                            Log.Infof("%v Time cost: %.2f ms ", start.Title, elapsed)
                    }
                    //格式化展示
                    if elapsed > 1000 {
                        elapsed = elapsed/1000
                        fmt.Printf("%vTime cost: %.2f s \n", start.Title, elapsed)
                    }else{
                        fmt.Printf("%vTime cost: %.2f ms \n", start.Title, elapsed)
                    }
                    
                } else {
                    //fmt.Println(fmt.Sprintf("%v:%v", to.RemoteAddr(), from.RemoteAddr()), " does not exist in the map.")
                }
			}
			
			// 设置超时时间
			to.SetWriteDeadline(time.Now().Add(timeOutSec))
			_, err = to.Write(buffer[:read])
// 			fmt.Println(string(buffer[5:read]))
			if err != nil {
				complete <- true
				oneSwitch <- true
				return
			}
		}
	}
}
