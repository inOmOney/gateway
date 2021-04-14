package http_proxy_middleware

import (
	"gateway/dao"
	"gateway/public"
	"github.com/gin-gonic/gin"
)

func HttpAppLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		appId, err := c.Get("app")
		if err == true {
			app := appId.(*dao.App)
			limiter := public.FLowLimitHandler.GetRateLimiter(app.AppID, int(app.Qps))
			if !limiter.Allow() {
				c.Writer.Write([]byte("租户限流"))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
