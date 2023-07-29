##go-upstream
服务代理中间件，支持后端服务集群配置，根据HASH配置选择后端机器进行代理。

##安装方法
```
go get github.com/fbbin/go-upstream
```

##Examples

```
bind: 0.0.0.0:9810
wait_queue_len: 10000
max_conn: 50
timeout: 60 #连接时长
failover: 3 #负载均衡尝试连接次数
type: "mysql" #mysql redis nginx 
stats: 0.0.0.0:8090
backend:
    - 127.0.1.1:3306
    - 127.0.1.1:3306
    - 127.0.1.1:3306

log:
    level: "info"
    path: "./logs/proxy.log"

```

```
[root@Web6 goup]# ./goup -c ./conf.yaml 
Start Proxy...
Start Successful...
2023/07/29 11:29:53 
2023/07/29 11:29:53 From client:192.168.5.254:36942 To Server:192.168.2.6:3307; Query: SET NAMES utf8
2023/07/29 11:29:59 From client:192.168.5.254:36942 To Server:192.168.2.6:3307; Quit: user quit

```



## http://1.1.1.1:8090/
```
Server connecting num:3 
Server:127.0.1.1:3306 FailTimes:0 isUp:true
```

## swoole and proxy swoole

```
[root@iZ2vc4fcja0fjd7ljf2a9cZ httpstatus]# ./httpstatus http://47.xx.xx.35

Connected to 47.xx.xx.35:80

Connected via plaintext

HTTP/1.1 200 OK
Server: nginx
Content-Type: text/html; charset=utf-8
Date: Thu, 07 Jul 2022 11:25:24 GMT
Vary: Accept-Encoding
Connection: keep-alive

Body discarded

   DNS Lookup   TCP Connection   Server Processing   Content Transfer
[       0ms  |           0ms  |              8ms  |             0ms  ]
             |                |                   |                  |
    namelookup:0ms            |                   |                  |
                        connect:0ms               |                  |
                                      starttransfer:9ms              |
                                                                 total:9ms      
[root@iZ2vc4fcja0fjd7ljf2a9cZ httpstatus]# 
[root@iZ2vc4fcja0fjd7ljf2a9cZ httpstatus]# ./httpstatus http://47.xx.xx.35:9810

Connected to 47.xxx.xx.35:9810

Connected via plaintext

HTTP/1.1 200 OK
Server: nginx
Content-Type: text/html; charset=utf-8
Date: Thu, 07 Jul 2022 11:25:25 GMT
Vary: Accept-Encoding
Connection: keep-alive

Body discarded

   DNS Lookup   TCP Connection   Server Processing   Content Transfer
[       0ms  |           0ms  |              9ms  |             0ms  ]
             |                |                   |                  |
    namelookup:0ms            |                   |                  |
                        connect:0ms               |                  |
                                      starttransfer:10ms             |
                                                                 total:10ms     

```
