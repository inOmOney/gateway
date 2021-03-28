package router

import (
	"gateway/controller"
	"gateway/lib"
	"gateway/log"
	"gateway/middleware"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//Redis引擎
	store, err := sessions.NewRedisStore(10, "tcp", lib.GetStringConf("base.session.redis_server"),
		"", []byte("secret"))
	if err != nil {
		log.Fatal("REDIS连接失败")
	}

	//登陆接口
	adminLoginRouter := router.Group("/admin_login")
	adminLoginRouter.Use(
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.TranslationMiddleware(),
		sessions.Sessions("mysession", store),
	)
	{
		controller.AdminLoginRegister(adminLoginRouter)
	}

	//管理员相关信息接口
	adminRouter := router.Group("/admin")
	adminRouter.Use(
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		sessions.Sessions("mysession", store),
		middleware.TranslationMiddleware(),
		middleware.SessionAuthMiddleware(), //验证管理员是否登陆
	)
	{
		controller.AdminRegister(adminRouter)
	}

	//service相关
	serviceRouter := router.Group("/service")
	serviceRouter.Use(
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		sessions.Sessions("mysession", store),
		middleware.TranslationMiddleware(),
		middleware.SessionAuthMiddleware(), //验证管理员是否登陆
	)
	{
		controller.ServiceRegister(serviceRouter)
	}


	//dashboard相关
	dashboardRouter := router.Group("/dashboard")
	dashboardRouter.Use(
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		sessions.Sessions("mysession", store),
		middleware.TranslationMiddleware(),
		middleware.SessionAuthMiddleware(), //验证管理员是否登陆
	)
	{
		controller.DashBoardRegister(dashboardRouter)
	}

	return router
}
