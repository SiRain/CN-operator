# 作业三

## 需求

- 构建本地镜像
- 编写 Dockerfile 将httpserver 容器化
- 将镜像推送至 docker 官方镜像仓库
- 通过 docker 命令本地启动 httpserver
- 通过 nsenter 进入容器查看 IP 配置

## httpserver代码

使用Go语言编写Http Server

```go
package main
 
import (
   "fmt"
   "log"
   "net"
   "net/http"
   "net/http/pprof"
   "os"
   "strings"
)
 
func index(w http.ResponseWriter, r *http.Request) {
   // w.Write([]byte("<h1>Welcome to Cloud Native</h1>"))
   // 03.设置version
   os.Setenv("VERSION", "v0.0.1")
   version := os.Getenv("VERSION")
   w.Header().Set("VERSION", version)
   fmt.Printf("os version: %s \n", version)
   // 02.将requst中的header 设置到 reponse中
   for k, v := range r.Header {
      for _, vv := range v {
         fmt.Printf("Header key: %s, Header value: %s \n", k, v)
         w.Header().Set(k, vv)
      }
   }
   // 04.记录日志并输出
   clientip := getCurrentIP(r)
   //fmt.Println(r.RemoteAddr)
   log.Printf("Success! Response code: %d", 200)
   log.Printf("Success! clientip: %s", clientip)
}
 
// 05.健康检查的路由
func healthz(w http.ResponseWriter, r *http.Request) {
   fmt.Fprintf(w, "working")
}
 
func getCurrentIP(r *http.Request) string {
   // 这里也可以通过X-Forwarded-For请求头的第一个值作为用户的ip，要注意的是这两个请求头代表的ip都有可能是伪造的
   ip := r.Header.Get("X-Real-IP")
   if ip == "" {
      // 当请求头不存在即不存在代理时直接获取ip
      ip = strings.Split(r.RemoteAddr, ":")[0]
   }
   return ip
}
 
// ClientIP 尽最大努力实现获取客户端 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx or haproxy）可以正常工作。
func ClientIP(r *http.Request) string {
   xForwardedFor := r.Header.Get("X-Forwarded-For")
   ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
   if ip != "" {
      return ip
   }
   ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
   if ip != "" {
      return ip
   }
   if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
      return ip
   }
   return ""
}
 
func main() {
   mux := http.NewServeMux()
   // 06. debug
   mux.HandleFunc("/debug/pprof/", pprof.Index)
   mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
   mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
   mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
   mux.HandleFunc("/", index)
   mux.HandleFunc("/healthz", healthz)
   if err := http.ListenAndServe(":8080", mux); err != nil {
      log.Fatalf("start http server failed, error: %s\n", err.Error())
   }
}
```

## 安装Go

```bash
root@yon:~# tar -C /usr/local -xzf go1.20.4.linux-amd64.tar.gz
[root@master go]# vim /etc/profile
# 新增如下
export PATH=$PATH:/usr/local/go/bin
[root@master go]# source !$
source /etc/profile
[root@master go]# go version
go version go1.20.4 linux/amd64
```

## 编译Go代码

先检查81端口占用情况：

```bash
root@yon:~/cncamp/module3# mkdir -p /root/cncamp/module3
root@yon:~/cncamp/module3# cd /root/cncamp/module3
root@yon:~/cncamp/module3# vim main.go
# 将httpserver代码写入
total 16
drwxr-xr-x 2 root root 4096 6月   5 19:59 ./
drwxr-xr-x 3 root root 4096 6月   5 19:56 ../
-rw-r--r-- 1 root root 2450 6月   5 19:58 main.go
root@yon:~/cncamp/module3# go build -o gohttpserver main.go
root@yon:~/cncamp/module3# ll
total 19
drwxr-xr-x 2 root root 4096 6月   5 19:59 ./
drwxr-xr-x 3 root root 4096 6月   5 19:56 ../
-rwxr-xr-x 1 root root 7676019 Jun  5 14:34 gohttpserver
-rw-r--r-- 1 root root    2373 Jun  5 09:54 main.go
# 启动httpserver
root@yon:~/cncamp/module3# ./gohttpserver
```

打开另一个终端，可以看到进程启动。

```bash
root@yon:~/cncamp/module3# netstat -atunlp | grep 8080
tcp6       0      0 :::8080                 :::*                    LISTEN      102382/./gohttpserv
```

打开浏览器访问如下地址

```
http://192.168.126.99:8080/healthz
```

显示文字working，说明编译成功

浏览器访问后，终端日志

```
os version: v0.0.1 
Header key: Connection, Header value: [keep-alive] 
Header key: User-Agent, Header value: [Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36] 
Header key: Accept, Header value: [image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8] 
Header key: Referer, Header value: [http://192.168.126.99:8080/healthz] 
Header key: Accept-Encoding, Header value: [gzip, deflate] 
Header key: Accept-Language, Header value: [zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7] 
2023/06/05 14:58:23 Success! Response code: 200
2023/06/05 14:58:23 Success! clientip: 192.168.126.1
```

## 注册Docker Hub

```
https://hub.docker.com/
```

## 构建镜像

### Dockerfile

```dockerfile
FROM golang:1.18 AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /build
COPY . .
RUN GOOS=linux  go build -installsuffix cgo -o httpserver main.go

FROM busybox
COPY --from=builder /build/httpserver /httpserver
EXPOSE 8080
ENTRYPOINT ["/httpserver"]
```

### 构建

```bash
root@yon:~/cncamp/module3# vim dockerfile 
# 将上一步的Dockerfile内容写入
root@yon:~/cncamp/module3# docker build -t httpserver:v1.0.1 . 
Sending build context to Docker daemon   5.12kB
Step 1/9 : FROM golang:1.18 AS builder
1.18: Pulling from library/golang
bbeef03cda1f: Pull complete 
f049f75f014e: Pull complete 
56261d0e6b05: Pull complete 
9bd150679dbd: Pull complete 
bfcb68b5bd10: Pull complete 
06d0c5d18ef4: Pull complete 
cc7973a07a5b: Pull complete 
Digest: sha256:50c889275d26f816b5314fc99f55425fa76b18fcaf16af255f5d57f09e1f48da
Status: Downloaded newer image for golang:1.18
 ---> c37a56a6d654
Step 2/9 : ENV GO111MODULE=on     CGO_ENABLED=0     GOPROXY=https://goproxy.cn,direct
 ---> Running in 4f51cea91a97
Removing intermediate container 4f51cea91a97
 ---> 574cc2154845
Step 3/9 : WORKDIR /build
 ---> Running in 95ac5b550e99
Removing intermediate container 95ac5b550e99
 ---> 549e75839b8e
Step 4/9 : COPY . .
 ---> 126526b03e33
Step 5/9 : RUN GOOS=linux  go build -installsuffix cgo -o httpserver main.go
 ---> Running in d3fa0c205eb6
Removing intermediate container d3fa0c205eb6
 ---> b02bfec73b3c
Step 6/9 : FROM busybox
latest: Pulling from library/busybox
325d69979d33: Downloading 
latest: Pulling from library/busybox
325d69979d33: Pull complete 
Digest: sha256:560af6915bfc8d7630e50e212e08242d37b63bd5c1ccf9bd4acccf116e262d5b
Status: Downloaded newer image for busybox:latest
 ---> 8135583d97fe
Step 7/9 : COPY --from=builder /build/httpserver /httpserver
 ---> 061256a3d9c6
Step 8/9 : EXPOSE 8080
 ---> Running in b7c4bab7abbb
Removing intermediate container b7c4bab7abbb
 ---> 018425edaddf
Step 9/9 : ENTRYPOINT ["/httpserver"]
 ---> Running in 514246d4eae5
Removing intermediate container 514246d4eae5
 ---> fc2b63f1b7eb
Successfully built fc2b63f1b7eb
Successfully tagged httpserver:v1.0.1
root@yon:~/cncamp/module3# docker images|grep httpserver
httpserver   v1.0.1    fc2b63f1b7eb   28 seconds ago       12.1MB
root@yon:~/cncamp/module3# docker run -d -p 81:8080 httpserver:v1.0.1
b05afc72189802a5ed77da663ba4f29479107772157557028ff7e094ffbd23ae
root@yon:~/cncamp/module3# docker ps
CONTAINER ID   IMAGE               COMMAND         CREATED         STATUS         PORTS                                   NAMES
b05afc721898   httpserver:v1.0.1   "/httpserver"   5 seconds ago   Up 4 seconds   0.0.0.0:81->8080/tcp, :::81->8080/tcp   infallible_austin
```

### 浏览器访问

```
http://192.168.126.99:81/healthz
```

返回working，说明容器构建成功

## 推送镜像至hub.docker.com

先创建新的仓库： yonchou/cncamp:tagname

官方给出了推送的命令：

```
docker tag local-image:tagname new-repo:tagname
docker push  yonchou/cncamp:tagname
```

先登录：

```bash
root@yon:~/cncamp/module3# docker login
Login with your Docker ID to push and pull images from Docker Hub. If you don't have a Docker ID, head over to https://hub.docker.com to create one.
Username: yonchou
Password: 
WARNING! Your password will be stored unencrypted in /root/.docker/config.json.
Configure a credential helper to remove this warning. See
https://docs.docker.com/engine/reference/commandline/login/#credentials-store

Login Succeeded
```

修改镜像名称并推送：

```bash
root@yon:~/cncamp/module3# docker images
REPOSITORY       TAG          IMAGE ID       CREATED          SIZE
httpserver       v1.0.1       fc2b63f1b7eb   19 minutes ago   12.1MB
<none>           <none>       b02bfec73b3c   19 minutes ago   1.03GB
busybox          latest       8135583d97fe   2 weeks ago      4.86MB
golang           1.18         c37a56a6d654   4 months ago     965MB
root@yon:~/cncamp/module3# docker tag httpserver:v1.0.1 yonchou/cncamp:v1.0.1
root@yon:~/cncamp/module3# docker images
REPOSITORY                  TAG       IMAGE ID       CREATED          SIZE
httpserver                  v1.0.1    fc2b63f1b7eb   33 minutes ago   12.1MB
yonchou/cncamp/httpserver   v1.0.1    fc2b63f1b7eb   33 minutes ago   12.1MB
yonchou/cncamp              v1.0.1    fc2b63f1b7eb   33 minutes ago   12.1MB
<none>                      <none>    b02bfec73b3c   33 minutes ago   1.03GB
busybox                     latest    8135583d97fe   2 weeks ago      4.86MB
golang                      1.18      c37a56a6d654   4 months ago     965MB

root@yon:~/cncamp/module3# docker rmi yonchou/cncamp/httpserver:v1.0.1
Untagged: yonchou/cncamp/httpserver:v1.0.1
root@yon:~/cncamp/module3# docker images
REPOSITORY                  TAG       IMAGE ID       CREATED          SIZE
httpserver                  v1.0.1    fc2b63f1b7eb   21 minutes ago   12.1MB
yonchou/cncamp/httpserver   v1.0.1    fc2b63f1b7eb   21 minutes ago   12.1MB
<none>                      <none>    b02bfec73b3c   21 minutes ago   1.03GB
busybox                     latest    8135583d97fe   2 weeks ago      4.86MB
golang                      1.18      c37a56a6d654   4 months ago     965MB

root@yon:~/cncamp/module3# docker push yonchou/cncamp
Using default tag: latest
The push refers to repository [docker.io/yonchou/cncamp]
tag does not exist: yonchou/cncamp:latest
root@yon:~/cncamp/module3# docker push yonchou/cncamp:v1.0.1
The push refers to repository [docker.io/yonchou/cncamp]
6ad7fc44849d: Pushed 
9547b4c33213: Pushed 
v1.0.1: digest: sha256:f0913d16ea31f3679de10b71dab75a3b458cdc5524b5ef1c907dda0001d0e029 size: 739
```

## 本地启动http server容器

```bash
root@yon:~/cncamp/module3# docker run -d -p 82:8080 yonchou/cncamp:v1.0.1
51105b852dbcea1865d4a92b1cd6b19a696863f07a18659c10b377d08c275998
```

和之前的状态相同，使用浏览器访问

```
http://192.168.126.99:82/healthz
```

返回working，说明容器正常。

## 进入容器查看 IP 配置

重启容器

```bash
root@yon:~/cncamp/module3# docker ps -a
CONTAINER ID   IMAGE                   COMMAND         CREATED              STATUS                      PORTS                                   NAMES
51105b852dbc   yonchou/cncamp:v1.0.1   "/httpserver"   About a minute ago   Up About a minute           0.0.0.0:82->8080/tcp, :::82->8080/tcp   xenodochial_haibt
b05afc721898   httpserver:v1.0.1       "/httpserver"   29 minutes ago       Up 29 minutes               0.0.0.0:81->8080/tcp, :::81->8080/tcp   infallible_austin
3a9795004fbd   httpserver:v1.0.1       "/httpserver"   33 minutes ago       Exited (2) 31 minutes ago                                           adoring_chebyshev
root@yon:~/cncamp/module3# docker stop b05afc721898
b05afc721898
root@yon:~/cncamp/module3# docker ps -a
CONTAINER ID   IMAGE                   COMMAND         CREATED          STATUS                      PORTS                                   NAMES
51105b852dbc   yonchou/cncamp:v1.0.1   "/httpserver"   2 minutes ago    Up 2 minutes                0.0.0.0:82->8080/tcp, :::82->8080/tcp   xenodochial_haibt
b05afc721898   httpserver:v1.0.1       "/httpserver"   30 minutes ago   Exited (2) 4 seconds ago                                            infallible_austin
3a9795004fbd   httpserver:v1.0.1       "/httpserver"   34 minutes ago   Exited (2) 31 minutes ago                                           adoring_chebyshev
root@yon:~/cncamp/module3# docker restart b05afc721898
b05afc721898
root@yon:~/cncamp/module3# docker ps -a
CONTAINER ID   IMAGE                   COMMAND         CREATED          STATUS                      PORTS                                   NAMES
51105b852dbc   yonchou/cncamp:v1.0.1   "/httpserver"   2 minutes ago    Up 2 minutes                0.0.0.0:82->8080/tcp, :::82->8080/tcp   xenodochial_haibt
b05afc721898   httpserver:v1.0.1       "/httpserver"   30 minutes ago   Up 3 seconds                0.0.0.0:81->8080/tcp, :::81->8080/tcp   infallible_austin
3a9795004fbd   httpserver:v1.0.1       "/httpserver"   34 minutes ago   Exited (2) 32 minutes ago                                           adoring_chebyshev
```


查看Docker镜像的PID：

```bash
root@yon:~/cncamp/module3# docker container top b05afc721898
UID                 PID                 PPID                C                   STIME               TTY                 TIME                CMD
root                10516               10491               0                   20:47               ?                   00:00:00            /httpserver
root@yon:~/cncamp/module3# docker inspect -f {{.State.Pid}} b05afc721898
10516
```

查看帮助：

```bash
root@yon:~/cncamp/module3# nsenter --help

Usage:
 nsenter [options] [<program> [<argument>...]]

Run a program with namespaces of other processes.

Options:
 -a, --all              enter all namespaces
 -t, --target <pid>     target process to get namespaces from
 -m, --mount[=<file>]   enter mount namespace
 -u, --uts[=<file>]     enter UTS namespace (hostname etc)
 -i, --ipc[=<file>]     enter System V IPC namespace
 -n, --net[=<file>]     enter network namespace
 -p, --pid[=<file>]     enter pid namespace
 -C, --cgroup[=<file>]  enter cgroup namespace
 -U, --user[=<file>]    enter user namespace
 -S, --setuid <uid>     set uid in entered namespace
 -G, --setgid <gid>     set gid in entered namespace
     --preserve-credentials do not touch uids or gids
 -r, --root[=<dir>]     set the root directory
 -w, --wd[=<dir>]       set the working directory
 -F, --no-fork          do not fork before exec'ing <program>
 -Z, --follow-context   set SELinux context according to --target PID

 -h, --help             display this help
 -V, --version          display version

For more details see nsenter(1).
```


根据PID进入Docker容器的网络命名空间：

```bash
root@yon:~/cncamp/module3# nsenter -n -t 10516
root@yon:~/cncamp/module3# ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
12: eth0@if13: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default 
    link/ether 02:42:ac:11:00:02 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 172.17.0.2/16 brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever
root@yon:~/cncamp/module3# exit
logout
```

查看网络配置：

```bash
root@yon:/run/docker/netns# cd /run/docker/netns
root@yon:/run/docker/netns# ll
total 0
drwxr-xr-x 2 root root  80 6月   5 20:47 ./
drwx------ 8 root root 180 6月   5 20:07 ../
-r--r--r-- 1 root root   0 6月   5 20:47 7b3206662c47
-r--r--r-- 1 root root   0 6月   5 20:44 ac6583dd7ef9
root@yon:/run/docker/netns# docker ps
CONTAINER ID   IMAGE                   COMMAND         CREATED          STATUS         PORTS                                   NAMES
51105b852dbc   yonchou/cncamp:v1.0.1   "/httpserver"   5 minutes ago    Up 5 minutes   0.0.0.0:82->8080/tcp, :::82->8080/tcp   xenodochial_haibt
b05afc721898   httpserver:v1.0.1       "/httpserver"   33 minutes ago   Up 3 minutes   0.0.0.0:81->8080/tcp, :::81->8080/tcp   infallible_austin
root@yon:/run/docker/netns# docker inspect b05afc721898| grep -i sandbox
            "SandboxID": "7b3206662c4705567928dbde00f6598a21c2fa37ccae8bd63b59724b58d1f0d7",
            "SandboxKey": "/var/run/docker/netns/7b3206662c47",
root@yon:/run/docker/netns# nsenter --net=/var/run/docker/netns/7b3206662c47 sh
# iptables -nvL -t mangle
Chain PREROUTING (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination         

Chain INPUT (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination         

Chain FORWARD (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination         

Chain OUTPUT (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination         

Chain POSTROUTING (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination         
# ipvsadm -ln
sh: 2: ipvsadm: not found
# exit
```


相关的namespace配置如下：

```bash
root@yon:/run/docker/netns# ll /proc/$(docker inspect -f {{.State.Pid}} b05afc721898)/ns
total 0
dr-x--x--x 2 root root 0 6月   5 20:47 ./
dr-xr-xr-x 9 root root 0 6月   5 20:47 ../
lrwxrwxrwx 1 root root 0 6月   5 20:51 cgroup -> 'cgroup:[4026531835]'
lrwxrwxrwx 1 root root 0 6月   5 20:51 ipc -> 'ipc:[4026532639]'
lrwxrwxrwx 1 root root 0 6月   5 20:51 mnt -> 'mnt:[4026532637]'
lrwxrwxrwx 1 root root 0 6月   5 20:47 net -> 'net:[4026532642]'
lrwxrwxrwx 1 root root 0 6月   5 20:51 pid -> 'pid:[4026532640]'
lrwxrwxrwx 1 root root 0 6月   5 20:51 pid_for_children -> 'pid:[4026532640]'
lrwxrwxrwx 1 root root 0 6月   5 20:51 time -> 'time:[4026531834]'
lrwxrwxrwx 1 root root 0 6月   5 20:51 time_for_children -> 'time:[4026531834]'
lrwxrwxrwx 1 root root 0 6月   5 20:51 user -> 'user:[4026531837]'
lrwxrwxrwx 1 root root 0 6月   5 20:51 uts -> 'uts:[4026532638]'
```

再次重启容器，可以看到PID变化，但容器IP并没有变化。

```bash
root@yon:/run/docker/netns# docker restart b05afc721898
b05afc721898
root@yon:/run/docker/netns# docker inspect -f {{.State.Pid}} b05afc721898
10748
root@yon:/run/docker/netns# nsenter -n -t `docker inspect -f {{.State.Pid}} b05afc721898`
root@yon:/run/docker/netns# ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
14: eth0@if15: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default 
    link/ether 02:42:ac:11:00:02 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 172.17.0.2/16 brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever
root@yon:/run/docker/netns# exit
logout
```
