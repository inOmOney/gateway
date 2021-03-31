package dao

import (
	"database/sql"
	"gateway/dto"
	"gateway/lib"
	"gateway/public"
	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

type ServiceInfo struct {
	ID          int64     `json:"id" save:"gateway_service_info,id"`
	LoadType    int       `json:"load_type" description:"负载类型 0=http 1=tcp 2=grpc"`
	ServiceName string    `json:"service_name" description:"服务名称"`
	ServiceDesc string    `json:"service_desc" description:"服务描述"`
	UpdateAt    time.Time `json:"create_at" description:"更新时间"`
	CreateAt    time.Time `json:"update_at" description:"添加时间"`
	IsDelete    int8      `json:"is_delete" description:"是否已删除；0：否；1：是"`
}

func (serviceInfo *ServiceInfo) ServiceNum(c *gin.Context, db *sql.DB)(int,error){
	table := "gateway_service_info"
	query := []string{"count(*) as serviceNum"}
	where := map[string]interface{}{
		"is_delete":    0,
	}
	cond, vals, err := builder.BuildSelect(table, where, query)
	if err!=nil {
		return 0, err
	}
	row, err := lib.DBQuery(public.GetGinTraceContext(c), db, cond, vals...)
	var ServiceNum int
	row.Next()
	err = row.Scan(&ServiceNum)
	if err!=nil {
		return 0,err
	}
	return ServiceNum,nil
}

func (serviceInfo *ServiceInfo) ServiceInfoPage(c *gin.Context, db *sql.DB, param *dto.ServiceListInput) ([]ServiceInfo, error) {

	// select * from gateway_service_info where service_name like %info% or service_desc like &info &&
	//		is_delete = 0 orderby id desc limit pageSize offset pageNo-1
	table := "gateway_service_info"
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
	var serviceInfos []ServiceInfo
	if err = scanner.Scan(rows, &serviceInfos); err != nil {
		return nil, err
	}
	return serviceInfos, err
}

// 查看service类型情况，按类型分组返回
func (serviceInfo *ServiceInfo) ServiceGroupByLoadType(c *gin.Context, db *sql.DB) ([]dto.DashServiceStatItemOutput, error){
	result := []dto.DashServiceStatItemOutput{}
	table := "gateway_service_info"
	query := []string{"load_type","count(*) as value"}
	where := map[string]interface{}{
		"is_delete":    0,
		"_groupby" : "load_type",
	}

	cond, vals, err := builder.BuildSelect(table, where, query)
	if err!=nil {
		return nil, err
	}
	row, err := lib.DBQuery(public.GetGinTraceContext(c), db, cond, vals...)
	scanner.Scan(row, &result)
	return result,nil
}

func (serviceInfo *ServiceInfo) ServiceDetail(c *gin.Context, db lib.TDManager) (*ServiceDetail, error) {

	info, err := serviceInfo.Find(c, db)
	if err != nil && err.Error() == "[scanner]: empty result" {
		return nil, errors.New("没有该条记录")
	}
	detail := &ServiceDetail{
		ServiceInfo: info,
	}
	httpRule := &HttpRule{ServiceID: serviceInfo.ID}
	httpRule, err = httpRule.Find(c, db) // 没有记录的错误处理
	if err != nil && err.Error() != "[scanner]: empty result" {
		return nil, err
	}
	tcp := &TcpRule{ServiceID: serviceInfo.ID}
	tcp, err = tcp.Find(c, db) // 没有记录的错误处理
	if err != nil && err.Error() != "[scanner]: empty result" {
		return nil, err
	}
	grpc := &GrpcRule{ServiceID: serviceInfo.ID}
	grpc, err = grpc.Find(c, db) // 没有记录的错误处理
	if err != nil && err.Error() != "[scanner]: empty result" {
		return nil, err
	}
	loadBalance := &LoadBalance{ServiceID: serviceInfo.ID}
	loadBalance, err = loadBalance.Find(c, db) // 没有记录的错误处理
	if err != nil && err.Error() != "[scanner]: empty result" {
		return nil, err
	}
	accessControl := &AccessControl{ServiceID: serviceInfo.ID}
	accessControl, err = accessControl.Find(c, db) // 没有记录的错误处理
	if err != nil && err.Error() != "[scanner]: empty result" {
		return nil, err
	}

	detail.HTTPRule = httpRule
	detail.TCPRule = tcp
	detail.GRPCRule = grpc
	detail.LoadBalance = loadBalance
	detail.AccessControl = accessControl

	return detail, nil

}

func (serviceInfo *ServiceInfo) Find(c *gin.Context, db lib.TDManager) (*ServiceInfo, error) {

	table := "gateway_service_info"
	query := []string{"*"}
	where := map[string]interface{}{
		"service_name": serviceInfo.ServiceName,
		"id":           serviceInfo.ID,
		"is_delete":    0,
	}
	finalWhere := builder.OmitEmpty(where, []string{"id", "service_name"})

	cond, vals, err := builder.BuildSelect(table, finalWhere, query)
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
	err = scanner.Scan(row, serviceInfo)
	return serviceInfo, err
}

func (serviceInfo *ServiceInfo) Update(c *gin.Context, db lib.TDManager) error {
	table := "gateway_service_info"
	data := map[string]interface{}{
		"load_type":    serviceInfo.LoadType,
		"service_name": serviceInfo.ServiceName,
		"service_desc": serviceInfo.ServiceDesc,
		"create_at":    serviceInfo.CreateAt,
		"update_at":    serviceInfo.UpdateAt,
		"is_delete":    serviceInfo.IsDelete,
	}
	finalData := builder.OmitEmpty(data, []string{"is_delete"})
	where := map[string]interface{}{
		"service_name": serviceInfo.ServiceName,
		"id":           serviceInfo.ID,
	}
	finalWhere := builder.OmitEmpty(where, []string{"service_name", "id"})
	cond, vals, err := builder.BuildUpdate(table, finalWhere, finalData)
	if err != nil {
		return err
	}
	result, err := lib.DBExec(public.GetGinTraceContext(c), db, cond, vals...)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	serviceInfo.ID = id
	return err
}
func (serviceInfo *ServiceInfo) Insert(c *gin.Context, db lib.TDManager) error {
	table := "gateway_service_info"
	data := []map[string]interface{}{
		map[string]interface{}{
			"load_type":    serviceInfo.LoadType,
			"service_name": serviceInfo.ServiceName,
			"service_desc": serviceInfo.ServiceDesc,
			"create_at":    time.Now(),
			"update_at":    time.Now(),
			"is_delete":    serviceInfo.IsDelete,
		},
	}

	cond, vals, err := builder.BuildInsert(table, data)
	if err != nil {
		return err
	}
	result, err := lib.DBExec(public.GetGinTraceContext(c), db, cond, vals...)

	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	serviceInfo.ID = id
	return err
}

func (serviceInfo *ServiceInfo) Save(c *gin.Context, db lib.TDManager) error {
	exist := serviceInfo.IsExist(c, db)
	if exist {
		return serviceInfo.Update(c, db)
	} else {
		return serviceInfo.Insert(c, db)
	}
}

func (serviceInfo *ServiceInfo) IsExist(c *gin.Context, db lib.TDManager) bool {
	table := "gateway_service_info"
	where := map[string]interface{}{
		"id":           serviceInfo.ID,
		"service_name": serviceInfo.ServiceName,
	}
	finalWhere := builder.OmitEmpty(where, []string{"id", "service_name"})
	what := []string{"*"}
	cond, vals, _ := builder.BuildSelect(table, finalWhere, what)
	_, err := db.Query(cond, vals)
	if err.Error() == "[scanner]: empty result" {
		return false
	} else {
		return true
	}
}

func (serviceInfo *ServiceInfo) Delete(c *gin.Context, db *sql.DB) error {

	table := "gateway_service_info"
	where := map[string]interface{}{
		"id": serviceInfo.ID,
	}
	update := map[string]interface{}{
		"is_delete": 1,
	}
	cond, vals, err := builder.BuildUpdate(table, where, update)
	if err != nil {
		return err
	}
	_, err = lib.DBExec(public.GetGinTraceContext(c), db, cond, vals...)
	return err

}
