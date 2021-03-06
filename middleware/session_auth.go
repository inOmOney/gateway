package middleware

import (
	"errors"
	"gateway/public"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if name, ok := session.Get(public.AdminSessionKey).(string); !ok || name == "" {
			ResponseError(c, UserNotLogin, errors.New("user not login"))
			c.Abort()
			return
		}
		c.Next()
	}
}
