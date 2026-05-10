package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type LoginPageData struct {
	Title        string
	TOTPRequired bool
	Error        string
	Lang         string
	I18n         I18n
	LoggedIn     bool
}

func (h *Handler) ShowLogin(c *gin.Context) {
	// 获取站点
	site := GetSiteByHost(c.Request.Host)
	if site == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	lang := detectLanguage(c.GetHeader("Accept-Language"))

	// 检查是否已登录
	loggedIn := false
	token, _ := c.Cookie(CookieName)
	if token != "" {
		_, err := ValidateToken(site.Config.JWTSecret, token)
		loggedIn = err == nil
	}

	data := LoginPageData{
		Title:        site.Config.Title,
		TOTPRequired: site.Auth.IsTOTPRequired(),
		Error:        c.Query("error"),
		Lang:         lang,
		I18n:         GetI18n(lang),
		LoggedIn:     loggedIn,
	}
	c.HTML(http.StatusOK, "login.html", data)
}

func (h *Handler) HandleLogin(c *gin.Context) {
	// 获取站点
	site := GetSiteByHost(c.Request.Host)
	if site == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	ip := c.ClientIP()
	lang := detectLanguage(c.GetHeader("Accept-Language"))
	i18n := GetI18n(lang)

	// 检查速率限制
	result := site.RateLimiter.Check(ip)
	if !result.Allowed {
		log.Printf("rate limited: site=%s, ip=%s, message=%s", site.Name, ip, result.Message)
		c.HTML(http.StatusTooManyRequests, "login.html", LoginPageData{
			Title:        site.Config.Title,
			TOTPRequired: site.Auth.IsTOTPRequired(),
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
	if err := site.Auth.Authenticate(username, password, totpCode); err != nil {
		site.RateLimiter.RecordFailure(ip)
		log.Printf("auth failed: site=%s, ip=%s, user=%s, error=%v", site.Name, ip, username, err)
		c.HTML(http.StatusUnauthorized, "login.html", LoginPageData{
			Title:        site.Config.Title,
			TOTPRequired: site.Auth.IsTOTPRequired(),
			Error:        i18n.ErrInvalid,
			Lang:         lang,
			I18n:         i18n,
		})
		return
	}

	// 认证成功，清除速率限制
	site.RateLimiter.RecordSuccess(ip)

	// 生成 JWT（使用该站点的 jwt_secret）
	token, err := GenerateToken(site.Config.JWTSecret, username, site.Config.SessionTTLHours)
	if err != nil {
		log.Printf("generate token error: %v", err)
		c.HTML(http.StatusInternalServerError, "login.html", LoginPageData{
			Title:        site.Config.Title,
			TOTPRequired: site.Auth.IsTOTPRequired(),
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
		site.Config.SessionTTLHours*3600,
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
	c.Redirect(http.StatusFound, "/inner-login")
}
