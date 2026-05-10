package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	config      *Config
	auth        *Authenticator
	rateLimiter *RateLimiter
}

func NewHandler(config *Config, auth *Authenticator, rateLimiter *RateLimiter) *Handler {
	return &Handler{
		config:      config,
		auth:        auth,
		rateLimiter: rateLimiter,
	}
}

type LoginPageData struct {
	Title        string
	TOTPRequired bool
	Error        string
}

func (h *Handler) ShowLogin(c *gin.Context) {
	data := LoginPageData{
		Title:        h.config.Title,
		TOTPRequired: h.auth.IsTOTPRequired(),
		Error:        c.Query("error"),
	}
	c.HTML(http.StatusOK, "login.html", data)
}

func (h *Handler) HandleLogin(c *gin.Context) {
	ip := c.ClientIP()

	// 检查速率限制
	result := h.rateLimiter.Check(ip)
	if !result.Allowed {
		log.Printf("rate limited: ip=%s, message=%s", ip, result.Message)
		c.HTML(http.StatusTooManyRequests, "login.html", LoginPageData{
			Title:        h.config.Title,
			TOTPRequired: h.auth.IsTOTPRequired(),
			Error:        result.Message,
		})
		return
	}

	// 获取表单数据
	username := c.PostForm("username")
	password := c.PostForm("password")
	totpCode := c.PostForm("totp")

	// 验证凭据
	if err := h.auth.Authenticate(username, password, totpCode); err != nil {
		h.rateLimiter.RecordFailure(ip)
		log.Printf("auth failed: ip=%s, user=%s, error=%v", ip, username, err)
		c.HTML(http.StatusUnauthorized, "login.html", LoginPageData{
			Title:        h.config.Title,
			TOTPRequired: h.auth.IsTOTPRequired(),
			Error:        "Invalid username or password",
		})
		return
	}

	// 认证成功，清除速率限制
	h.rateLimiter.RecordSuccess(ip)

	// 生成 JWT
	token, err := GenerateToken(h.config.JWTSecret, username, h.config.SessionTTLHours)
	if err != nil {
		log.Printf("generate token error: %v", err)
		c.HTML(http.StatusInternalServerError, "login.html", LoginPageData{
			Title:        h.config.Title,
			TOTPRequired: h.auth.IsTOTPRequired(),
			Error:        "Internal server error",
		})
		return
	}

	// 设置 Cookie
	c.SetCookie(
		CookieName,
		token,
		h.config.SessionTTLHours*3600,
		"/",
		"",
		false, // secure, 生产环境应设为 true
		true,  // httpOnly
	)

	// 重定向到原始请求路径或首页
	redirectTo := c.Query("redirect")
	if redirectTo == "" {
		redirectTo = "/"
	}
	c.Redirect(http.StatusFound, redirectTo)
}

func (h *Handler) HandleLogout(c *gin.Context) {
	c.SetCookie(CookieName, "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}
