package main

import (
	"net/http/httputil"
	"strings"
	"time"
)

type Site struct {
	Name        string
	Config      *SiteConfig
	Auth        *Authenticator
	RateLimiter *RateLimiter
	Proxy       *httputil.ReverseProxy
}

var sitesByHost map[string]*Site

func InitSites(config *Config) error {
	sitesByHost = make(map[string]*Site)

	for name, siteConfig := range config.Sites {
		// 创建反向代理
		proxy, err := NewReverseProxy(siteConfig.Upstream)
		if err != nil {
			return err
		}

		// 创建站点实例
		site := &Site{
			Name:        name,
			Config:      &siteConfig,
			Auth:        NewAuthenticator(&siteConfig.Auth),
			RateLimiter: NewRateLimiter(&siteConfig.RateLimit),
			Proxy:       proxy,
		}

		// 存储到 map，使用 listen_host 作为 key
		sitesByHost[siteConfig.ListenHost] = site
	}

	return nil
}

func GetSiteByHost(host string) *Site {
	// 移除端口号
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return sitesByHost[host]
}

func CleanupAllRateLimiters() {
	for _, site := range sitesByHost {
		site.RateLimiter.Cleanup()
	}
}

func StartCleanupTicker() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			CleanupAllRateLimiters()
		}
	}()
}
