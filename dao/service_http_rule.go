package dao

import (
	"gateway/lib"
	"gateway/public"
	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	"github.com/gin-gonic/gin"
)

type HttpRule struct {
	ID             int64  `json:"id" gorm:"primary_key"`
	ServiceID      int64  `json:"service_id" gorm:"column:service_id" description:"服务id"`
	RuleType       int    `json:"rule_type" gorm:"column:rule_type" description:"匹配类型 domain=域名, url_prefix=url前缀"`
	Rule           string `json:"rule" gorm:"column:rule" description:"type=domain表示域名，type=url_prefix时表示url前缀"`
	NeedHttps      int    `json:"need_https" gorm:"column:need_https" description:"type=支持https 1=支持"`
	NeedWebsocket  int    `json:"need_websocket" gorm:"column:need_websocket" description:"启用websocket 1=启用"`
	NeedStripUri   int    `json:"need_strip_uri" gorm:"column:need_strip_uri" description:"启用strip_uri 1=启用"`
	UrlRewrite     string `json:"url_rewrite" gorm:"column:url_rewrite" description:"url重写功能，每行一个	"`
	HeaderTransfor string `json:"header_transfor" gorm:"column:header_transfor" description:"header转换支持增加(add)、删除(del)、修改(edit) 格式: add headname headvalue	"`
}

func (httpRule *HttpRule) Find(c *gin.Context, db lib.TDManager) (*HttpRule, error) {
	table := "gateway_service_http_rule"
	query := []string{"*"}
	where := map[string]interface{}{
		"service_id": httpRule.ServiceID,
	}

	cond, vals, err := builder.BuildSelect(table, where, query)
	if err != nil {
		return nil, err
	}

	row, err := lib.DBQuery(public.GetGinTraceContext(c), db, cond, vals...)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	err = scanner.Scan(row, httpRule)
	return httpRule, err
}

func (httpRule *HttpRule) Save(c *gin.Context, db lib.TDManager) error {
	exist := httpRule.IsExist(c, db)
	if exist {
		return httpRule.Update(c, db)
	} else {
		return httpRule.Insert(c, db)
	}
}

func (httpRule *HttpRule) IsExist(c *gin.Context, db lib.TDManager) bool {
	table := "gateway_service_http_rule"
	where := map[string]interface{}{
		"service_id": httpRule.ServiceID,
	}
	what := []string{"*"}
	cond, vals, _ := builder.BuildSelect(table, where, what)
	_, err := lib.DBQuery(public.GetGinTraceContext(c), db, cond, vals...)

	if err.Error() == "[scanner]: empty result" {
		return false
	} else {
		return true
	}
}

func (httpRule *HttpRule) Update(c *gin.Context, db lib.TDManager) error {
	table := "gateway_service_http_rule"
	data := map[string]interface{}{
		"need_https":      httpRule.NeedHttps,
		"need_strip_uri":  httpRule.NeedStripUri,
		"need_websocket":  httpRule.NeedWebsocket,
		"url_rewrite":     httpRule.UrlRewrite,
		"header_transfor": httpRule.HeaderTransfor,
	}
	finalData := builder.OmitEmpty(data, []string{"need_https", "need_strip_uri", "need_websocket", "url_rewrite", "header_transfor"})
	where := map[string]interface{}{
		"service_id": httpRule.ServiceID,
	}
	cond, vals, err := builder.BuildUpdate(table, where, finalData)
	if err != nil {
		return err
	}
	_, err = lib.DBExec(public.GetGinTraceContext(c), db, cond, vals...)
	return err
}

func (httpRule *HttpRule) Insert(c *gin.Context, db lib.TDManager) error {
	table := "gateway_service_http_rule"
	data := []map[string]interface{}{
		map[string]interface{}{
			"service_id":      httpRule.ServiceID,
			"rule_type":       httpRule.RuleType,
			"rule":            httpRule.Rule,
			"need_https":      httpRule.NeedHttps,
			"need_strip_uri":  httpRule.NeedStripUri,
			"need_websocket":  httpRule.NeedWebsocket,
			"url_rewrite":     httpRule.UrlRewrite,
			"header_transfor": httpRule.HeaderTransfor,
		},
	}
	cond, vals, err := builder.BuildInsert(table, data)
	if err != nil {
		return err
	}
	_, err = lib.DBExec(public.GetGinTraceContext(c), db, cond, vals...)
	return err
}
