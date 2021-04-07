package controller

import (
	"gateway/dao"
	"gateway/dto"
	"gateway/lib"
	"gateway/middleware"
	"gateway/public"
	"github.com/gin-gonic/gin"
)

func APPRegister(g *gin.RouterGroup) {
	g.Handle("GET", "/app_list", AppList)
}

func AppList(c *gin.Context) {
	params := &dto.APPListInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	app := &dao.App{}
	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	apps, err := app.FindByPage(c, params, db)
	if err!=nil {
		middleware.ResponseError(c, 2003, err)
		return
	}


	out := &dto.APPListOutput{}
	for _, item := range apps {
		counter := public.FlowCountHandler.GetServiceCountHandler(item.AppID, c)

		out.List = append(out.List, dto.APPListItemOutput{
			ID:       item.ID,
			AppID:    item.AppID,
			Name:     item.Name,
			Secret:   item.Secret,
			WhiteIPS: item.WhiteIPS,
			Qpd:      item.Qpd,
			Qps:      item.Qps,
			RealQpd: counter.DayTotal,
			RealQps: counter.Qps,
		} )
	}
	out.Total = len(apps)
	middleware.ResponseSuccess(c, out)
}
