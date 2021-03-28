package dao

import (
	"gateway/lib"
	"gateway/public"
	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	"github.com/gin-gonic/gin"
)

type GrpcRule struct {
	ID             int64  `json:"id" gorm:"primary_key"`
	ServiceID      int64  `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	Port           int    `json:"port" gorm:"column:port" description:"端口	"`
	HeaderTransfor string `json:"header_transfor" gorm:"column:header_transfor" description:"header转换支持增加(add)、删除(del)、修改(edit) 格式: add headname headvalue"`
}

func (grpc *GrpcRule) Find(c *gin.Context, db lib.TDManager) (*GrpcRule, error) {
	table := "gateway_service_grpc_rule"
	query := []string{"*"}
	where := map[string]interface{}{
		"service_id": grpc.ServiceID,
		"port":       grpc.Port,
	}
	finalWhere := builder.OmitEmpty(where, []string{"service_id", "port"})

	cond, vals, err := builder.BuildSelect(table, finalWhere, query)
	if err != nil {
		return nil, err
	}

	row, err := lib.DBQuery(public.GetGinTraceContext(c), db, cond, vals...)
	if err != nil {
		return nil, err
	}
	err = scanner.Scan(row, grpc)
	return grpc, err
}

func (grpc *GrpcRule) Update(c *gin.Context, db lib.TDManager) error {
	table := "gateway_service_grpc_rule"
	data := map[string]interface{}{
		"header_transfor": grpc.HeaderTransfor,
	}
	where := map[string]interface{}{
		"service_id": grpc.ServiceID,
	}
	cond, vals, err := builder.BuildUpdate(table, where, data)
	if err != nil {
		return err
	}
	_, err = lib.DBExec(public.GetGinTraceContext(c), db, cond, vals...)
	return err
}

func (grpc *GrpcRule) Insert(c *gin.Context, db lib.TDManager) error {
	table := "gateway_service_grpc_rule"
	data := []map[string]interface{}{
		map[string]interface{}{
			"service_id": grpc.ServiceID,
			"port":       grpc.Port,
			"header_transfor": grpc.HeaderTransfor,
		},
	}
	cond, vals, err := builder.BuildInsert(table, data)
	if err != nil {
		return err
	}
	_, err = lib.DBExec(public.GetGinTraceContext(c), db, cond, vals...)
	return err
}
