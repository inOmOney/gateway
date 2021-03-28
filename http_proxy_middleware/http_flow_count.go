package http_proxy_middleware

import (
	"gateway/dao"
	"gateway/public"
	"github.com/gin-gonic/gin"
)

func HttpFlowCount() gin.HandlerFunc{
	return func(c *gin.Context) {
		service, exists := c.Get("service")
		if !exists {
			panic("没有设置对应的服务")
		}
		// 拿到统计抓手 字段+1
		global := public.FlowCountHandler.GetServiceCountHandler(public.GlobalFlowCount, c)
		global.Increase()

		detail := service.(*dao.ServiceDetail)
		// 拿到统计抓手 字段+1
		handler := public.FlowCountHandler.GetServiceCountHandler(detail.ServiceInfo.ServiceName, c)
		handler.Increase()

		c.Next()
	}
}
