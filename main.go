package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

const version = "0.0.2"

func main() {
	showVersion := flag.Bool("v", false, "print version")
	showTOTP := flag.String("show-totp", "", "show TOTP import URI for specified site name")
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

	if *showTOTP != "" {
		siteConfig, ok := config.Sites[*showTOTP]
		if !ok {
			log.Fatalf("site %s not found", *showTOTP)
		}
		if siteConfig.Auth.TOTPToken == "" {
			log.Fatalf("site %s: totp_token is not configured", *showTOTP)
		}
		uri := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
			siteConfig.Title, siteConfig.Auth.User, siteConfig.Auth.TOTPToken, siteConfig.Title)
		fmt.Println(uri)
		return
	}

	// 初始化所有站点
	if err := InitSites(config); err != nil {
		log.Fatalf("init sites: %v", err)
	}

	// 启动定时清理速率限制记录
	StartCleanupTicker()

	// 设置 Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	// 创建 handler
	handler := NewHandler()

	// 静态路由
	r.GET("/inner-login", handler.ShowLogin)
	r.POST("/inner-login", handler.HandleLogin)
	r.GET("/inner-logout", handler.HandleLogout)

	// 其他路由需要认证
	r.NoRoute(AuthMiddleware(), func(c *gin.Context) {
		site := GetSiteByHost(c.Request.Host)
		if site == nil {
			c.AbortWithStatus(404)
			return
		}
		site.Proxy.ServeHTTP(c.Writer, c.Request)
	})

	// 启动服务
	addr := fmt.Sprintf(":%d", config.ListenPort)
	log.Printf("inner-auth starting on %s", addr)
	for host, site := range sitesByHost {
		log.Printf("site: %s -> %s", host, site.Config.Upstream)
	}
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
