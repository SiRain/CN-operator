package main

import (
	"fmt"
	"os"
	"strconv"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	// 功能1
	http.HandleFunc("/requestAndResponse", requestAndResponse)
	// 功能2
	http.HandleFunc("/getVersion", getVersion)
	// 功能3
	http.HandleFunc("/ipAndStatus", ipAndStatus)
	// 功能4
    http.HandleFunc("/healthz", healthz)

	err := http.ListenAndServe(":8080", nil) 
	if nil != err {
		log.Fatal(err) 
	}
}

// 功能1：接收请求，处理header后返回响应
func requestAndResponse(response http.ResponseWriter, request *http.Request) {
	headers := request.Header // header是Map类型数据
	for header := range headers {  // value是[]string切片
		values := headers[header]
		for index, _ := range values {
			values[index] = strings.TrimSpace(values[index])
		}
		println(header + "=" + strings.Join(values, ","))        // 打印request的header的k=v
		response.Header().Set(header, strings.Join(values, ",")) // 遍历写入response的Header
	}
	fmt.Fprintln(response, "Header全部数据:", headers)
	io.WriteString(response, "succeed")
}

// 功能2：获取环境变量的version
func getVersion(response http.ResponseWriter, request *http.Request) {
	envStr := os.Getenv("VERSION")
	response.Header().Set("VERSION", envStr)
	io.WriteString(response, "succeed")
}

// 功能3：输出IP与返回码
func ipAndStatus(response http.ResponseWriter, request *http.Request) {
	form := request.RemoteAddr
	println("Client->ip:port=" + form)
	ipStr := strings.Split(form, ":")
	println("Client->ip=" + ipStr[0])  // 打印ip
	println("Client->response code=" + strconv.Itoa(http.StatusOK))
	io.WriteString(response, "succeed")
}

// 功能4：连通性测试接口
func healthz(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(200)  // 设置返回码200
	io.WriteString(response, "succeed")
}