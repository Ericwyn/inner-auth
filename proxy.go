package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type ReverseProxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

func NewReverseProxy(upstream string) (*ReverseProxy, error) {
	target, err := url.Parse(upstream)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Header.Set("X-Forwarded-Proto", "http")
		// 修改 Origin header 以匹配上游服务
		if origin := req.Header.Get("Origin"); origin != "" {
			req.Header.Set("Origin", target.Scheme+"://"+target.Host)
		}
	}

	// 保留所有 header，不删除 hop-by-hop header
	proxy.FlushInterval = -1

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	return &ReverseProxy{
		target: target,
		proxy:  proxy,
	}, nil
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if isWebSocketRequest(r) {
		rp.serveWebSocket(w, r)
		return
	}
	rp.proxy.ServeHTTP(w, r)
}

func isWebSocketRequest(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Connection"), "upgrade") &&
		strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func (rp *ReverseProxy) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	// 建立到上游的连接
	targetAddr := rp.target.Host
	if !strings.Contains(targetAddr, ":") {
		targetAddr += ":80"
	}

	backendConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("websocket dial error: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer backendConn.Close()

	// 构建转发给上游的请求
	outReq := new(http.Request)
	*outReq = *r
	outReq.URL.Scheme = rp.target.Scheme
	outReq.URL.Host = rp.target.Host
	outReq.Host = rp.target.Host
	outReq.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	outReq.Header.Set("X-Forwarded-Proto", "http")
	// 修改 Origin header 以匹配上游服务
	if origin := outReq.Header.Get("Origin"); origin != "" {
		outReq.Header.Set("Origin", rp.target.Scheme+"://"+rp.target.Host)
	}

	// 将请求写入上游连接
	if err := outReq.Write(backendConn); err != nil {
		log.Printf("websocket write request error: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	// Hijack 客户端连接
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("websocket hijack not supported")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("websocket hijack error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// 双向转发数据
	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}

	go cp(backendConn, clientConn)
	go cp(clientConn, backendConn)

	// 等待任一方向完成
	<-errc
}
