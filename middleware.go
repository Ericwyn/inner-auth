package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否有 token
		token, err := c.Cookie(CookieName)
		if err != nil || token == "" {
			// 无 token，重定向到登录页
			redirectTo := c.Request.URL.Path
			c.Redirect(http.StatusFound, "/inner-login?redirect="+redirectTo)
			c.Abort()
			return
		}

		// 验证 token
		claims, err := ValidateToken(jwtSecret, token)
		if err != nil {
			// token 无效，重定向到登录页
			redirectTo := c.Request.URL.Path
			c.Redirect(http.StatusFound, "/inner-login?redirect="+redirectTo)
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user", claims.User)
		c.Next()
	}
}
