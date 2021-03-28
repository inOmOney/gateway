package http_proxy_router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"gateway/http_proxy_middleware"
)

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine{
	r := gin.New()
	r.Use(middlewares...)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK,gin.H{
			"message":"pong",
		})
	})
	r.Use(
		http_proxy_middleware.HttpAccessModeMiddleware(),
		http_proxy_middleware.HttpFlowCount(),
		http_proxy_middleware.HttpReverseProxyMiddleware(),
	)
	return r
}