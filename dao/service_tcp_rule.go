package dao

import (
	"gateway/lib"
	"gateway/public"
	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	"github.com/gin-gonic/gin"
)

type TcpRule struct {
	ID        int64 `json:"id" gorm:"primary_key"`
	ServiceID int64 `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	Port      int   `json:"port" gorm:"column:port" description:"端口	"`
}

func (tcpRule *TcpRule) Find(c *gin.Context, db lib.TDManager) (*TcpRule, error) {
	table := "gateway_service_tcp_rule"
	query := []string{"*"}
	where := map[string]interface{}{
		"service_id": tcpRule.ServiceID,
		"port":       tcpRule.Port,
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
	err = scanner.Scan(row, tcpRule)
	return tcpRule, err
}

func (tcpRule *TcpRule) Insert(c *gin.Context, db lib.TDManager) error {
	table := "gateway_service_tcp_rule"
	data := []map[string]interface{}{
		map[string]interface{}{
			"service_id": tcpRule.ServiceID,
			"port":       tcpRule.Port,
		},
	}
	cond, vals, err := builder.BuildInsert(table, data)
	if err != nil {
		return err
	}
	_, err = lib.DBExec(public.GetGinTraceContext(c), db, cond, vals...)
	return err
}
