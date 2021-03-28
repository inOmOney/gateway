package http_proxy_middleware

import (
	"gateway/dao"
	"gateway/reverse_proxy"
	"github.com/gin-gonic/gin"
)

func HttpReverseProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		service, exists := c.Get("service")
		if !exists {
			panic("没有设置对应的服务")
		}
		detail := service.(*dao.ServiceDetail)
		lb := dao.LoadBalancerHandler.GetLoadBalancer(detail)
		proxy := reverse_proxy.GetReverseProxy(c, lb)
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
		return
	}
}
