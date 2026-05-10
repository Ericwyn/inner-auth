package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取站点
		site := GetSiteByHost(c.Request.Host)
		if site == nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// 检查是否有 token
		token, err := c.Cookie(CookieName)
		if err != nil || token == "" {
			// 无 token，重定向到登录页
			redirectTo := c.Request.URL.Path
			c.Redirect(http.StatusFound, "/inner-login?redirect="+redirectTo)
			c.Abort()
			return
		}

		// 验证 token（使用该站点的 jwt_secret）
		claims, err := ValidateToken(site.Config.JWTSecret, token)
		if err != nil {
			// token 无效，重定向到登录页
			redirectTo := c.Request.URL.Path
			c.Redirect(http.StatusFound, "/inner-login?redirect="+redirectTo)
			c.Abort()
			return
		}

		// 将用户信息和站点信息存入上下文
		c.Set("user", claims.User)
		c.Set("site", site)
		c.Next()
	}
}
