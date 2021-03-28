package http_proxy_middleware

import (
	"errors"
	"gateway/dao"
	"gateway/middleware"
	"github.com/gin-gonic/gin"
)

func HttpAccessModeMiddleware() gin.HandlerFunc{
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/favicon.ico" {
			middleware.ResponseError(c, 1001, errors.New("没有logo"))
			c.Abort()
			return
		}

		detail, err := dao.SvcManager.GetDetailFromReq(c)
		if err!=nil {
			middleware.ResponseError(c, 1001, err)
			c.Abort()
			return
		}
		c.Set("service", detail)
		c.Next()
	}
}
