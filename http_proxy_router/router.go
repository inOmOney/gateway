package http_proxy_router

import (
	"gateway/controller"
	"gateway/http_proxy_middleware"
	"gateway/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(middlewares...)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	oauth := r.Group("/oauth")
	oauth.Use(
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.TranslationMiddleware(),
	)
	controller.OAuthRegister(oauth)
	//todo 限流
	r.Use(
		http_proxy_middleware.HttpAccessModeMiddleware(),
		http_proxy_middleware.HttpFlowCount(),
		http_proxy_middleware.HttpAppJwtValid(),
		http_proxy_middleware.HttpAppFlowCount(),
		//http_proxy_middleware.HttpAppLimit(),
		http_proxy_middleware.HttpStripUri(),
		http_proxy_middleware.HttpRewriteUrl(),
		http_proxy_middleware.HttpReverseProxyMiddleware(),
	)

	return r
}
