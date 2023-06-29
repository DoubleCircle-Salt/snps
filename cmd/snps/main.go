package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

var (
	requestMap = make(map[string]net.Conn)
	l  sync.RWMutex
)

var (
	connChan = make(chan net.Conn, 128)
)

func copy(conn1, conn2 net.Conn, errChan chan<- error) {
	_, err := io.Copy(conn1, conn2)
	errChan <- err
}

var (
	ErrInvalidWrite = errors.New("invalid write result")
)

func copyConn(dst, src net.Conn, errChan chan<- error) {
	var err error
	buf := make([]byte, 4096)
	for {
		src.SetReadDeadline(time.Now().Add(time.Second))
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = ErrInvalidWrite
				}
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			//if er != io.EOF {
				err = er
			//}
			break
		}
	}
	errChan <- err
}

func serve(conn net.Conn) {
	defer conn.Close()

	bufioReader := bufio.NewReader(conn)

	_, err := http.ReadRequest(bufioReader)
	if err != nil {
		println("read request failed, err:", err.Error())
		return
	}

	id := uuid.NewV4()

	responseHeader := make(http.Header)
	responseHeader.Add("Address", "http://113.31.159.165:8888/" + id.String())

	response := &http.Response{
		StatusCode: 200,
		Header:     responseHeader,
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	if err := response.Write(conn); err != nil {
		println("response write failed, err:", err.Error())
		return
	}

	l.Lock()
	requestMap[id.String()] = conn
	l.Unlock()


	for {
		conn2 := <- connChan
		errChan := make(chan error, 2)
		go copyConn(conn, conn2, errChan)
		go copyConn(conn2, conn, errChan)

		err := <- errChan
		if err != nil {
			println("copy err:", err.Error())
		}
		conn2.Close()
	}
}

func listenProxy() {
	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		println("listen failed, err:", err.Error())
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			println("accept failed, err:", err.Error())
			return
		}
		go func(conn net.Conn) {
			connChan <- conn
		}(conn)
	}
}

func main() {

	go listenProxy()
	
	ln, err := tls.Listen("tcp", ":7777", &tls.Config{
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			return defaultCert, nil
		},
	})
	if err != nil {
		println("listen failed, err:", err.Error())
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			println("accept failed, err:", err.Error())
			return
		}

		go serve(conn)
	}
}
