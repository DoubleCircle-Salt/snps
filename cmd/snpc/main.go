package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
)

func handler(ctx *fasthttp.RequestCtx) {
	fasthttp.ServeFile(ctx, "./Makefile")
}

func getAddress(address string) {

	resp, err := http.Get(address)
	if err != nil {
		println("http get failed, err:", err.Error())
		return
	}

	println("status:", resp.StatusCode)

	if resp.Body == nil {
		return
	}

	buf := make([]byte, 4096)

	for {
		n, err := resp.Body.Read(buf)
		if err != nil {
			println("body read failed, err:", err.Error())
			return
		}
		fmt.Println("body:")
		fmt.Println(string(buf[:n]))
	}
}

func main() {

	dialer := &net.Dialer{Timeout: 3 * time.Second}

	conn, err := tls.DialWithDialer(dialer, "tcp", "113.31.159.165:7777", &tls.Config{
		ServerName: "yyy.test.com",
		InsecureSkipVerify: true,
	})
	if err != nil {
		println("dial failed, err:", err.Error())
		return
	}
	defer conn.Close()

	// 请求上车点，获取下载地址
	request, err := http.NewRequest("GET", "https://yyy.test.com:7777", nil)
	if err != nil {
		println("new request failed, err:", err.Error())
		return
	}

	request.Header.Add("Token", "123456")

	if err := request.Write(conn); err != nil {
		println("write request failed, err:", err.Error())
		return
	}

	bufioReader := bufio.NewReader(conn)

	response, err := http.ReadResponse(bufioReader, request)
	if err != nil {
		println("read response failed, err:", err.Error())
		return
	}

	address := response.Header.Get("Address")
	println("address:", address)

	go func() {
		time.Sleep(3 * time.Second)
		// 请求文件示例
		getAddress(address)
	}()
	// 创建文件服务
	srv := &fasthttp.Server{
	}
	srv.Handler = handler
	err = srv.ServeConn(conn)
	println("serve err:", err)

	for{}
}
