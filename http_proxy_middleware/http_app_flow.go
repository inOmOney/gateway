package http_proxy_middleware

import (
	"fmt"
	"gateway/dao"
	"gateway/lib"
	"gateway/log"
	"gateway/public"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
)

//以租户为单位进行流量统计
func HttpAppFlowCount() gin.HandlerFunc {
	return func(c *gin.Context) {
		appId, ok := c.Get("app")
		if ok == true {
			app, _ := appId.(*dao.App)
			handler := public.FlowCountHandler.GetServiceCountHandler(app.AppID, c)
			defer handler.Increase() //每次请求对内存中进行周期内访问数量统计，定期推送给redis
			if app.Qpd > 0 {
				rdb, _ := lib.RedisConnFactory("default")
				dayKey := handler.GetDayKey()
				fmt.Println(dayKey)
				result, err := redis.Int64(lib.RedisLogDo(public.GetGinTraceContext(c), rdb, "GET", dayKey))
				if err != nil {
					log.Error("redis读取数据出错")
					c.Abort()
					return
				}
				if result >= app.Qpd {
					c.Writer.Write([]byte("租户该日流量已用尽"))
					c.Abort()
					return
				}
			}
		}
		c.Next()
	}
}
