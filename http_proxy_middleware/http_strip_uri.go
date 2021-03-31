package http_proxy_middleware

import (
	"gateway/dao"
	"gateway/public"
	"github.com/gin-gonic/gin"
	"strings"
)

func HttpStripUri() gin.HandlerFunc {
	return func(c *gin.Context) {
		service, exists := c.Get("service")
		if !exists {
			panic("没有设置对应的服务")
		}
		detail, _ := service.(*dao.ServiceDetail)

		if detail.ServiceInfo.LoadType == public.HttpLoadType &&detail.HTTPRule.NeedStripUri == 1 && detail.HTTPRule.RuleType == public.UrlRuleType{
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, detail.HTTPRule.Rule,"",1)
		}
		c.Next()
	}
}
