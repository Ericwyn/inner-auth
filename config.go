package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	ListenPort int                   `json:"listen_port"`
	Sites      map[string]SiteConfig `json:"sites"`
}

type SiteConfig struct {
	ListenHost      string          `json:"listen_host"`
	Upstream        string          `json:"upstream"`
	Title           string          `json:"title"`
	JWTSecret       string          `json:"jwt_secret"`
	SessionTTLHours int             `json:"session_ttl_hours"`
	CookieSecure    *bool           `json:"cookie_secure"`
	RateLimit       RateLimitConfig `json:"rate_limit"`
	Auth            AuthConfig      `json:"auth"`
}

type RateLimitConfig struct {
	MaxAttemptsPerIP    int `json:"max_attempts_per_ip"`
	IPWindowSeconds     int `json:"ip_window_seconds"`
	IPLockoutSeconds    int `json:"ip_lockout_seconds"`
	GlobalMaxAttempts   int `json:"global_max_attempts"`
	GlobalWindowSeconds int `json:"global_window_seconds"`
}

type AuthConfig struct {
	User         string `json:"user"`
	Password     string `json:"password"`
	PasswordHash string `json:"password_hash"`
	TOTPToken    string `json:"totp_token"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	// 设置默认值
	if cfg.ListenPort == 0 {
		cfg.ListenPort = 6904
	}

	// 验证至少有一个站点
	if len(cfg.Sites) == 0 {
		return nil, fmt.Errorf("at least one site is required")
	}

	// 验证每个站点的必填项
	for name, site := range cfg.Sites {
		if site.ListenHost == "" {
			return nil, fmt.Errorf("site %s: listen_host is required", name)
		}
		if site.Upstream == "" {
			return nil, fmt.Errorf("site %s: upstream is required", name)
		}
		if site.JWTSecret == "" {
			return nil, fmt.Errorf("site %s: jwt_secret is required", name)
		}
		if site.Auth.User == "" {
			return nil, fmt.Errorf("site %s: auth.user is required", name)
		}
		if site.Auth.Password == "" && site.Auth.PasswordHash == "" {
			return nil, fmt.Errorf("site %s: auth.password or auth.password_hash is required", name)
		}

		// 设置站点默认值
		if site.Title == "" {
			site.Title = "Login"
		}
		if site.SessionTTLHours == 0 {
			site.SessionTTLHours = 168
		}
		if site.RateLimit.MaxAttemptsPerIP == 0 {
			site.RateLimit.MaxAttemptsPerIP = 5
		}
		if site.RateLimit.IPWindowSeconds == 0 {
			site.RateLimit.IPWindowSeconds = 60
		}
		if site.RateLimit.IPLockoutSeconds == 0 {
			site.RateLimit.IPLockoutSeconds = 300
		}
		if site.RateLimit.GlobalMaxAttempts == 0 {
			site.RateLimit.GlobalMaxAttempts = 200
		}
		if site.RateLimit.GlobalWindowSeconds == 0 {
			site.RateLimit.GlobalWindowSeconds = 3600
		}

		cfg.Sites[name] = site
	}

	return cfg, nil
}
