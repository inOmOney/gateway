package dao

import (
	"gateway/lib"
	"gateway/public"
	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"sync"
	"time"
)

type AppManager struct{
	AppMap map[string]*App // key:app_id, value:secret
	Init sync.Once

}
var AppManagerHandler *AppManager

func init(){
	AppManagerHandler = &AppManager{
		AppMap: map[string]*App{},
		Init: sync.Once{},
	}

}
func (m *AppManager) LoadOnceAppInfo(){
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	db, _ := lib.GetDBPool("default")
	param := &APPListInput{
		PageSize: 9999,
		PageNo: 1,
	}
	app := &App{}
	apps, err := app.FindByPage(c, param, db)
	if err!=nil {
		panic(err)
	}
	for index, value := range apps {
		AppManagerHandler.AppMap[apps[index].AppID] = &value
	}
}

func (app *App) FindByPage(c *gin.Context, param *APPListInput, db lib.TDManager) ([]App,error) {
	table := "gateway_app"
	var where map[string]interface{}
	if  param.Info == ""   {
		where = map[string]interface{}{
			"is_delete =": 0,
			"_orderby":    "id desc",
			"_limit":      []uint{uint(param.PageNo - 1), uint(param.PageSize)},
		}
	} else {
		where = map[string]interface{}{
			"_or": []map[string]interface{}{
				{"service_name like": "%" + param.Info + "%"},
				{"service_desc like": "%" + param.Info + "%"},
			},
			"is_delete =": 0,
			"_orderby":    "id desc",
			"_limit":      []uint{uint(param.PageNo - 1), uint(param.PageSize)},
		}
	}

	query := []string{"*"}
	cond, vals, err := builder.BuildSelect(table, where, query)
	if err != nil {
		return nil, err
	}
	rows, err := lib.DBQuery(public.GetGinTraceContext(c), db, cond, vals...)
	if err != nil {
		return nil, err
	}
	var a []App
	if err = scanner.Scan(rows, &a); err != nil {
		return nil, err
	}
	return a, nil
}


type APPListInput struct {
	Info     string `json:"info" form:"info" comment:"查找信息" validate:""`
	PageSize int    `json:"page_size" form:"page_size" comment:"页数" validate:"required,min=1,max=999"`
	PageNo   int    `json:"page_no" form:"page_no" comment:"页码" validate:"required,min=1,max=999"`
}
type App struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	AppID     string    `json:"app_id" gorm:"column:app_id" description:"租户id	"`
	Name      string    `json:"name" gorm:"column:name" description:"租户名称	"`
	Secret    string    `json:"secret" gorm:"column:secret" description:"密钥"`
	WhiteIPS  string    `json:"white_ips" gorm:"column:white_ips" description:"ip白名单，支持前缀匹配"`
	Qpd       int64     `json:"qpd" gorm:"column:qpd" description:"日请求量限制"`
	Qps       int64     `json:"qps" gorm:"column:qps" description:"每秒请求量限制"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"添加时间	"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}
