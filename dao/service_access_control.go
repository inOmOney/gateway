package dao

import (
	"gateway/lib"
	"gateway/public"
	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	"github.com/gin-gonic/gin"
)

type AccessControl struct {
	ID                int64  `json:"id" gorm:"primary_key"`
	ServiceID         int64  `json:"service_id" gorm:"column:service_id" description:"服务id"`
	OpenAuth          int    `json:"open_auth" gorm:"column:open_auth" description:"是否开启权限 1=开启"`
	BlackList         string `json:"black_list" gorm:"column:black_list" description:"黑名单ip	"`
	WhiteList         string `json:"white_list" gorm:"column:white_list" description:"白名单ip	"`
	WhiteHostName     string `json:"white_host_name" gorm:"column:white_host_name" description:"白名单主机	"`
	ClientIPFlowLimit int    `json:"clientip_flow_limit" gorm:"column:clientip_flow_limit" description:"客户端ip限流	"`
	ServiceFlowLimit  int    `json:"service_flow_limit" gorm:"column:service_flow_limit" description:"服务端限流	"`
}

func (accessControl *AccessControl) Find(c *gin.Context, db lib.TDManager) (*AccessControl, error) {
	table := "gateway_service_access_control"
	query := []string{"*"}
	where := map[string]interface{}{
		"service_id": accessControl.ServiceID,
	}

	cond, vals, err := builder.BuildSelect(table, where, query)
	if err != nil {
		return nil, err
	}

	row, err := lib.DBQuery(public.GetGinTraceContext(c), db, cond, vals...)
	if err != nil {
		return nil, err
	}
	err = scanner.Scan(row, accessControl)
	return accessControl, err
}

func (accessControl *AccessControl) Save(c *gin.Context, db lib.TDManager) error {
	exist := accessControl.IsExist(c, db)
	if exist {
		return accessControl.Update(c, db)
	} else {
		return accessControl.Insert(c, db)
	}
}
func (accessControl *AccessControl) IsExist(c *gin.Context, db lib.TDManager) bool{
	table := "gateway_service_access_control"
	where := map[string]interface{}{
		"service_id":accessControl.ServiceID,
	}
	what := []string{"*"}
	cond, vals, _ := builder.BuildSelect(table, where, what)
	_, err := db.Query(cond, vals)
	if err.Error() == "[scanner]: empty result" {
		return false
	} else {
		return true
	}
}

func (accessControl *AccessControl) Update(c *gin.Context, db lib.TDManager) error {
	table := "gateway_service_access_control"

	data := map[string]interface{}{
		"open_auth":           accessControl.OpenAuth,
		"black_list":          accessControl.BlackList,
		"white_list":          accessControl.WhiteList,
		"clientip_flow_limit": accessControl.ClientIPFlowLimit,
		"service_flow_limit":  accessControl.ServiceFlowLimit,
	}
	where := map[string]interface{}{
		"service_id": accessControl.ServiceID,
	}


	cond, vals, err := builder.BuildUpdate(table, where, data)
	if err != nil {
		return err
	}
	_, err = lib.DBExec(public.GetGinTraceContext(c), db, cond, vals...)
	return err
}
func (accessControl *AccessControl) Insert(c *gin.Context, db lib.TDManager) error {
	table := "gateway_service_access_control"
	data := []map[string]interface{}{
		map[string]interface{}{
			"service_id": 	accessControl.ServiceID,
			"open_auth":           accessControl.OpenAuth,
			"black_list":          accessControl.BlackList,
			"white_list":          accessControl.WhiteList,
			"clientip_flow_limit": accessControl.ClientIPFlowLimit,
			"service_flow_limit":  accessControl.ServiceFlowLimit,
		},
	}

	cond, vals, err := builder.BuildInsert(table, data)
	if err != nil {
		return err
	}
	_, err = lib.DBExec(public.GetGinTraceContext(c), db, cond, vals...)
	return err
}
