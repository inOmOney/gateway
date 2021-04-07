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

	g.GET("/panel_group_data", service.GetPanelGroupData) //返回当日总访问，当前qps，服务数
}

func (o *DashBoardController) GetPanelGroupData(c *gin.Context) {
	handler := public.FlowCountHandler.GetServiceCountHandler(public.GlobalFlowCount, c)
	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	num, err := serviceInfo.ServiceNum(c, db)
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	app := &dao.App{}
	count, err := app.TotalCount(c, db)
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	out := &dto.DashPanelOutput{
		CurrentQps:      handler.Qps,
		TodayRequestNum: handler.DayTotal,
		ServiceNum:      num,
		AppNum: count,
	}
	middleware.ResponseSuccess(c, out)
}

func (o *DashBoardController) GetServiceStat(c *gin.Context) {
	serviceInfo := &dao.ServiceInfo{}
	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	item, err := serviceInfo.ServiceGroupByLoadType(c, db)

	out := &dto.DashServiceStatOutput{
		Data:   item,
		Legend: []string{},
	}
	for index := range item {
		if item[index].LoadType == public.HttpLoadType {
			item[index].Name = "HTTP"
		} else if item[index].LoadType == public.TcpLoadType {
			item[index].Name = "TCP"
		} else if item[index].LoadType == public.GrpcLoadType {
			item[index].Name = "GRPC"
		}
		out.Legend = append(out.Legend, item[index].Name)
	}
	middleware.ResponseSuccess(c, out)
}
func (o *DashBoardController) GetFlow(c *gin.Context) {
	redisConn, err := lib.RedisConnFactory("default")
	//defer redisConn.Close()
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
