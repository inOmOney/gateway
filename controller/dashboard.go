package controller

import (
	"gateway/dao"
	"gateway/dto"
	"gateway/lib"
	"gateway/middleware"
	"gateway/public"
	"github.com/gin-gonic/gin"
)

type DashBoardController struct{}

func DashBoardRegister(g *gin.RouterGroup) {
	service := &DashBoardController{}
	g.GET("/flow_stat", service.GetFlow)
	g.GET("/service_stat", service.GetServiceStat)
}

func (o *DashBoardController) GetServiceStat(c *gin.Context) {
	serviceInfo := &dao.ServiceInfo{}
	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
	}

	item, err := serviceInfo.ServiceGroupByLoadType(c, db)

	out := &dto.DashServiceStatOutput{
		Data : item,
		Legend: []string{},
	}
	for index := range item {
		if item[index].LoadType == public.HttpLoadType{
			item[index].Name = "HTTP"
		}else if item[index].LoadType == public.TcpLoadType {
			item[index].Name = "TCP"
		}else if item[index].LoadType == public.GrpcLoadType {
			item[index].Name = "GRPC"
		}
		out.Legend = append(out.Legend, item[index].Name)
	}
	middleware.ResponseSuccess(c, out)
}
func (o *DashBoardController) GetFlow(c *gin.Context) {
	redisConn, err := lib.RedisConnFactory("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err) // 获取redis连接失败
		return
	}
	today, err := public.GetTodayFlow(public.GlobalFlowCount, redisConn, c)
	yesterday, err := public.GetYesterdayFlow(public.GlobalFlowCount, redisConn, c)

	out := &dto.ServiceStatOutput{
		Today:     today,
		Yesterday: yesterday,
	}
	middleware.ResponseSuccess(c, out)
}
