package main

import (
	"sync"
	"time"
)

type IPRecord struct {
	Attempts    int
	LastAttempt time.Time
	LockedUntil time.Time
}

type RateLimiter struct {
	mu                sync.RWMutex
	config            *RateLimitConfig
	ipRecords         map[string]*IPRecord
	globalAttempts    int
	globalWindowStart time.Time
	globalLocked      bool
	globalLockedUntil time.Time
}

func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		config:            config,
		ipRecords:         make(map[string]*IPRecord),
		globalWindowStart: time.Now(),
	}
}

type RateLimitResult struct {
	Allowed    bool
	Locked     bool
	RetryAfter time.Duration
	Message    string
}

func (rl *RateLimiter) Check(ip string) RateLimitResult {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 检查全局限制
	if rl.globalLocked {
		if time.Now().Before(rl.globalLockedUntil) {
			retryAfter := time.Until(rl.globalLockedUntil)
			return RateLimitResult{
				Allowed:    false,
				Locked:     true,
				RetryAfter: retryAfter,
				Message:    "System busy, please try again later",
			}
		}
		// 全局锁定已过期，重置
		rl.globalLocked = false
		rl.globalAttempts = 0
		rl.globalWindowStart = time.Now()
	}

	// 检查全局限制窗口
	if time.Since(rl.globalWindowStart) > time.Duration(rl.config.GlobalWindowSeconds)*time.Second {
		rl.globalAttempts = 0
		rl.globalWindowStart = time.Now()
	}

	if rl.globalAttempts >= rl.config.GlobalMaxAttempts {
		rl.globalLocked = true
		// 锁定到当前小时结束
		now := time.Now()
		rl.globalLockedUntil = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
		retryAfter := time.Until(rl.globalLockedUntil)
		return RateLimitResult{
			Allowed:    false,
			Locked:     true,
			RetryAfter: retryAfter,
			Message:    "System busy, please try again later",
		}
	}

	// 检查 IP 限制
	record := rl.ipRecords[ip]
	if record != nil {
		// 检查 IP 锁定
		if time.Now().Before(record.LockedUntil) {
			retryAfter := time.Until(record.LockedUntil)
			return RateLimitResult{
				Allowed:    false,
				Locked:     true,
				RetryAfter: retryAfter,
				Message:    "Too many attempts, please try again later",
			}
		}

		// 检查 IP 窗口
		if time.Since(record.LastAttempt) > time.Duration(rl.config.IPWindowSeconds)*time.Second {
			record.Attempts = 0
		}

		if record.Attempts >= rl.config.MaxAttemptsPerIP {
			record.LockedUntil = time.Now().Add(time.Duration(rl.config.IPLockoutSeconds) * time.Second)
			retryAfter := time.Until(record.LockedUntil)
			return RateLimitResult{
				Allowed:    false,
				Locked:     true,
				RetryAfter: retryAfter,
				Message:    "Too many attempts, please try again later",
			}
		}
	}

	return RateLimitResult{Allowed: true}
}

func (rl *RateLimiter) RecordFailure(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 增加全局计数
	rl.globalAttempts++

	// 增加 IP 计数
	record := rl.ipRecords[ip]
	if record == nil {
		record = &IPRecord{}
		rl.ipRecords[ip] = record
	}
	record.Attempts++
	record.LastAttempt = time.Now()
}

func (rl *RateLimiter) RecordSuccess(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 清除该 IP 的记录
	delete(rl.ipRecords, ip)
}

func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, record := range rl.ipRecords {
		// 清理过期的 IP 记录
		if now.Sub(record.LastAttempt) > time.Duration(rl.config.IPLockoutSeconds)*time.Second {
			delete(rl.ipRecords, ip)
		}
	}
}
