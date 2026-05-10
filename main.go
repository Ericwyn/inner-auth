package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

const version = "0.0.1"

func main() {
	showVersion := flag.Bool("v", false, "print version")
	showTOTP := flag.Bool("show-totp", false, "show TOTP import URI")
	configPath := flag.String("c", "config.json", "config file path")
	flag.Parse()

	if *showVersion {
		fmt.Printf("inner-auth %s\n", version)
		return
	}

	// 加载配置
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if *showTOTP {
		if config.Auth.TOTPToken == "" {
			log.Fatal("totp_token is not configured")
		}
		uri := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
			config.Title, config.Auth.User, config.Auth.TOTPToken, config.Title)
		fmt.Println(uri)
		return
	}

	// 初始化组件
	auth := NewAuthenticator(&config.Auth)
	rateLimiter := NewRateLimiter(&config.RateLimit)
	handler := NewHandler(config, auth, rateLimiter)

	// 创建反向代理
	proxy, err := NewReverseProxy(config.Upstream)
	if err != nil {
		log.Fatalf("create proxy: %v", err)
	}

	// 启动定时清理速率限制记录
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rateLimiter.Cleanup()
		}
	}()

	// 设置 Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	// 静态路由
	r.GET("/login", handler.ShowLogin)
	r.POST("/login", handler.HandleLogin)
	r.GET("/logout", handler.HandleLogout)

	// 其他路由需要认证
	r.NoRoute(AuthMiddleware(config.JWTSecret), func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	// 启动服务
	addr := fmt.Sprintf(":%d", config.ListenPort)
	log.Printf("inner-auth starting on %s", addr)
	log.Printf("upstream: %s", config.Upstream)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
