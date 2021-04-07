package http_proxy_middleware

import (
	"gateway/public"
	"github.com/gin-gonic/gin"
)


//以租户为单位进行流量统计
func HttpAppFlowCount()gin.HandlerFunc{
	return func(c *gin.Context) {
		appId, _ := c.Get("app")
		appIdStr, _ := appId.(string)
		handler := public.FlowCountHandler.GetServiceCountHandler(appIdStr, c)
		handler.Increase()
	}
}