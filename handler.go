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
	Lang         string
	I18n         I18n
}

func (h *Handler) ShowLogin(c *gin.Context) {
	lang := detectLanguage(c.GetHeader("Accept-Language"))
	data := LoginPageData{
		Title:        h.config.Title,
		TOTPRequired: h.auth.IsTOTPRequired(),
		Error:        c.Query("error"),
		Lang:         lang,
		I18n:         GetI18n(lang),
	}
	c.HTML(http.StatusOK, "login.html", data)
}

func (h *Handler) HandleLogin(c *gin.Context) {
	ip := c.ClientIP()
	lang := detectLanguage(c.GetHeader("Accept-Language"))
	i18n := GetI18n(lang)

	// 检查速率限制
	result := h.rateLimiter.Check(ip)
	if !result.Allowed {
		log.Printf("rate limited: ip=%s, message=%s", ip, result.Message)
		c.HTML(http.StatusTooManyRequests, "login.html", LoginPageData{
			Title:        h.config.Title,
			TOTPRequired: h.auth.IsTOTPRequired(),
			Error:        i18n.ErrRateLimit,
			Lang:         lang,
			I18n:         i18n,
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
			Error:        i18n.ErrInvalid,
			Lang:         lang,
			I18n:         i18n,
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
			Error:        i18n.ErrInternal,
			Lang:         lang,
			I18n:         i18n,
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
