package http_proxy_middleware

import (
	"gateway/dao"
	"github.com/gin-gonic/gin"
	"regexp"
	"strings"
)

func HttpRewriteUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		service, exists := c.Get("service")
		if !exists {
			panic("没有设置对应的服务")
		}
		detail, _ := service.(*dao.ServiceDetail)
		tempPath := []byte(c.Request.URL.Path)
		for _, value := range strings.Split(detail.HTTPRule.UrlRewrite, ",") { //   ^/aaa/(.*) /bbb/$1
			zz := strings.Split(value, " ")
			com, err := regexp.Compile(zz[0])
			if err != nil {
				panic(err)
			}
			tempPath = com.ReplaceAll(tempPath, []byte(zz[1]))
		}
		c.Request.URL.Path = string(tempPath)
		c.Next()
	}
}
