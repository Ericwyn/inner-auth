package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	ListenPort     int          `json:"listen_port"`
	Upstream       string       `json:"upstream"`
	Title          string       `json:"title"`
	JWTSecret      string       `json:"jwt_secret"`
	SessionTTLHours int         `json:"session_ttl_hours"`
	RateLimit      RateLimitConfig `json:"rate_limit"`
	Auth           AuthConfig   `json:"auth"`
}

type RateLimitConfig struct {
	MaxAttemptsPerIP  int `json:"max_attempts_per_ip"`
	IPWindowSeconds   int `json:"ip_window_seconds"`
	IPLockoutSeconds  int `json:"ip_lockout_seconds"`
	GlobalMaxAttempts int `json:"global_max_attempts"`
	GlobalWindowSeconds int `json:"global_window_seconds"`
}

type AuthConfig struct {
	User      string `json:"user"`
	Password  string `json:"password"`
	TOTPToken string `json:"totp_token"`
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
	if cfg.Title == "" {
		cfg.Title = "Login"
	}
	if cfg.SessionTTLHours == 0 {
		cfg.SessionTTLHours = 168
	}
	if cfg.RateLimit.MaxAttemptsPerIP == 0 {
		cfg.RateLimit.MaxAttemptsPerIP = 5
	}
	if cfg.RateLimit.IPWindowSeconds == 0 {
		cfg.RateLimit.IPWindowSeconds = 60
	}
	if cfg.RateLimit.IPLockoutSeconds == 0 {
		cfg.RateLimit.IPLockoutSeconds = 300
	}
	if cfg.RateLimit.GlobalMaxAttempts == 0 {
		cfg.RateLimit.GlobalMaxAttempts = 200
	}
	if cfg.RateLimit.GlobalWindowSeconds == 0 {
		cfg.RateLimit.GlobalWindowSeconds = 3600
	}

	// 验证必填项
	if cfg.Upstream == "" {
		return nil, fmt.Errorf("upstream is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("jwt_secret is required")
	}
	if cfg.Auth.User == "" {
		return nil, fmt.Errorf("auth.user is required")
	}
	if cfg.Auth.Password == "" {
		return nil, fmt.Errorf("auth.password is required")
	}

	return cfg, nil
}
