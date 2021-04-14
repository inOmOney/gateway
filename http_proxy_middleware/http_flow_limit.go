package http_proxy_middleware

import (
	"fmt"
	"gateway/dao"
	"gateway/public"
	"github.com/gin-gonic/gin"
	"log"
)

func HttpFlowLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		service, exists := c.Get("service")
		if !exists {
			panic("没有设置对应的服务")
		}
		detail := service.(*dao.ServiceDetail)

		if detail.AccessControl.ServiceFlowLimit > 0 {
			serviceLimiter := public.FLowLimitHandler.GetRateLimiter(detail.ServiceInfo.ServiceName, detail.AccessControl.ServiceFlowLimit)
			if !serviceLimiter.Allow() {
				c.Writer.Write([]byte("服务端限流"))
				c.Abort()
				return
			}
		}

		if detail.AccessControl.ClientIPFlowLimit > 0 {
			clientLimiter := public.FLowLimitHandler.GetRateLimiter(
				fmt.Sprintf("%s_%s", detail.ServiceInfo.ServiceName, c.ClientIP()),
				detail.AccessControl.ClientIPFlowLimit)

			if !clientLimiter.Allow() {
				c.Writer.Write([]byte("客户端限流"))
				log.Print("fail")
				c.Abort()
				return
			}
		}
		log.Print("success")
		c.Next()
	}

}
