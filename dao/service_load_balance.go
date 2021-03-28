package dao

import (
	"gateway/lib"
	"gateway/public"
	"gateway/reverse_proxy/load_balance"
	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	"github.com/gin-gonic/gin"
	"sync"
)

type LoadBalance struct {
	ID            int64  `json:"id" gorm:"primary_key"`
	ServiceID     int64  `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	CheckMethod   int    `json:"check_method" gorm:"column:check_method" description:"检查方法 tcpchk=检测端口是否握手成功	"`
	CheckTimeout  int    `json:"check_timeout" gorm:"column:check_timeout" description:"check超时时间	"`
	CheckInterval int    `json:"check_interval" gorm:"column:check_interval" description:"检查间隔, 单位s		"`
	RoundType     int    `json:"round_type" gorm:"column:round_type" description:"轮询方式 round/weight_round/random/ip_hash"`
	IpList        string `json:"ip_list" gorm:"column:ip_list" description:"ip列表"`
	WeightList    string `json:"weight_list" gorm:"column:weight_list" description:"权重列表"`
	ForbidList    string `json:"forbid_list" gorm:"column:forbid_list" description:"禁用ip列表"`

	UpstreamConnectTimeout int `json:"upstream_connect_timeout" gorm:"column:upstream_connect_timeout" description:"下游建立连接超时, 单位s"`
	UpstreamHeaderTimeout  int `json:"upstream_header_timeout" gorm:"column:upstream_header_timeout" description:"下游获取header超时, 单位s	"`
	UpstreamIdleTimeout    int `json:"upstream_idle_timeout" gorm:"column:upstream_idle_timeout" description:"下游链接最大空闲时间, 单位s	"`
	UpstreamMaxIdle        int `json:"upstream_max_idle" gorm:"column:upstream_max_idle" description:"下游最大空闲链接数"`
}

func (loadBalance *LoadBalance) Find(c *gin.Context, db lib.TDManager) (*LoadBalance, error) {
	table := "gateway_service_load_balance"
	query := []string{"*"}
	where := map[string]interface{}{
		"service_id": loadBalance.ServiceID,
	}

	cond, vals, err := builder.BuildSelect(table, where, query)
	if err != nil {
		return nil, err
	}

	row, err := lib.DBQuery(public.GetGinTraceContext(c), db, cond, vals...)
	if err != nil {
		return nil, err
	}
	row.Scan()
	err = scanner.Scan(row, loadBalance)
	return loadBalance, err
}

func (loadBalance *LoadBalance) Update(c *gin.Context, db lib.TDManager) error {
	table := "gateway_service_load_balance"
	data := map[string]interface{}{
		"service_id":  loadBalance.ServiceID,
		"round_type":  loadBalance.RoundType,
		"ip_list":     loadBalance.IpList,
		"weight_list": loadBalance.WeightList,
		"forbid_list": loadBalance.ForbidList,
	}
	where := map[string]interface{}{
		"service_id": loadBalance.ServiceID,
	}
	cond, vals, err := builder.BuildUpdate(table, where, data)
	if err != nil {
		return err
	}
	_, err = lib.DBExec(public.GetGinTraceContext(c), db, cond, vals...)
	return err
}

func (loadBalance *LoadBalance) Insert(c *gin.Context, db lib.TDManager) error {
	table := "gateway_service_load_balance"

	data := []map[string]interface{}{
		map[string]interface{}{
			"service_id":  loadBalance.ServiceID,
			"round_type":  loadBalance.RoundType,
			"ip_list":     loadBalance.IpList,
			"weight_list": loadBalance.WeightList,
			"forbid_list": loadBalance.ForbidList,
		},
	}

	cond, vals, err := builder.BuildInsert(table, data)
	if err != nil {
		return err
	}
	_, err = lib.DBExec(public.GetGinTraceContext(c), db, cond, vals...)
	return err

}

var LoadBalancerHandler *LoadBalancer

type LoadBalancer struct {
	LoadBalanceMap   map[string]load_balance.LoadBalance
	LoadBalanceSlice []load_balance.LoadBalance
	Locker           sync.RWMutex
}

func init() {
	LoadBalancerHandler = &LoadBalancer{
		LoadBalanceMap:   map[string]load_balance.LoadBalance{},
		LoadBalanceSlice: []load_balance.LoadBalance{},
		Locker:           sync.RWMutex{},
	}
}



func (handler *LoadBalancer)GetLoadBalancer(detail *ServiceDetail)load_balance.LoadBalance{

	// 内存中查找 没有则查数据库
	if lb, ok := handler.LoadBalanceMap[detail.ServiceInfo.ServiceName]; ok {
		return lb
	}
	//
	schema := ""
	if detail.ServiceInfo.LoadType == public.HttpLoadType  {
		schema = "http://"
	}
	config := load_balance.NewConfig(schema, detail.ServiceInfo.ServiceName, detail.LoadBalance.IpList, detail.LoadBalance.WeightList )
	loadBalanceInstance := load_balance.LBFactory(load_balance.LBType(detail.LoadBalance.RoundType), config)
	handler.Locker.Lock()
	defer handler.Locker.Unlock()
	handler.LoadBalanceMap[detail.ServiceInfo.ServiceName] = loadBalanceInstance
	return loadBalanceInstance
}


