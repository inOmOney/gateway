package controller

import (
	"gateway/dao"
	"gateway/dto"
	"gateway/lib"
	"gateway/middleware"
	"gateway/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

type ServiceController struct{}

func ServiceRegister(g *gin.RouterGroup) {
	service := &ServiceController{}
	g.GET("/service_list", service.ServiceList)
	g.POST("/service_add_http", service.ServiceAddHttp)
	g.POST("/service_update_http", service.ServiceUpdateHttp)
	g.GET("/service_detail", service.ServiceDetail)
	g.GET("/service_delete", service.ServiceDelete)

	g.POST("/service_add_tcp", service.ServiceAddTcp)
	g.POST("/service_update_tcp", service.ServiceUpdateTcp)
	g.POST("/service_add_grpc", service.ServiceAddGrpc)
	g.POST("/service_update_grpc", service.ServiceUpdateGrpc)
}

func (service *ServiceController) ServiceList(c *gin.Context) {
	/*
		1. 绑参数
		2. 分页查serviceInfo
		3. 遍历每个serviceInfo 根据serviceId 配置output
	*/
	param := &dto.ServiceListInput{}
	if err := param.BindValidInput(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	serviceInfos, err := serviceInfo.ServiceInfoPage(c, db, param)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	//遍历每个info
	outList := []dto.ServiceListItemOutput{}
	for _, listItem := range serviceInfos {

		detail, err := listItem.ServiceDetail(c, db)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			return
		}
		//服务地址
		serviceAddr := "nil"
		clusterIP := lib.GetStringConf("base.cluster.cluster_ip")            //"127.0.0.1"
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")        //8080
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port") //4433
		if detail.ServiceInfo.LoadType == public.HttpLoadType &&
			detail.HTTPRule.RuleType == public.UrlRuleType &&
			detail.HTTPRule.NeedHttps == public.NeedHttps {
			serviceAddr = clusterIP + ":" + clusterSSLPort + detail.HTTPRule.Rule
		} else if detail.ServiceInfo.LoadType == public.HttpLoadType &&
			detail.HTTPRule.RuleType == public.UrlRuleType &&
			detail.HTTPRule.NeedHttps == public.NotNeedHttps {
			serviceAddr = clusterIP + ":" + clusterPort + detail.HTTPRule.Rule
		} else if detail.ServiceInfo.LoadType == public.HttpLoadType &&
			detail.HTTPRule.RuleType == public.DomainRuleType {
			serviceAddr = detail.HTTPRule.Rule
		} else if detail.ServiceInfo.LoadType == public.TcpLoadType {
			serviceAddr = clusterIP + ":" + strconv.Itoa(detail.TCPRule.Port)
		} else if detail.ServiceInfo.LoadType == public.GrpcLoadType {
			serviceAddr = clusterIP + ":" + strconv.Itoa(detail.GRPCRule.Port)
		}

		//节点数
		loadBalance := &dao.LoadBalance{ServiceID: listItem.ID}
		outLoadBalance, err := loadBalance.Find(c, db)
		ipList := strings.Split(outLoadBalance.IpList, ",")

		//todo qps 和qpd
		counter := public.FlowCountHandler.GetServiceCountHandler(listItem.ServiceName, c)
		//汇总
		outItem := dto.ServiceListItemOutput{
			ID:          listItem.ID,
			LoadType:    listItem.LoadType,
			ServiceName: listItem.ServiceName,
			ServiceDesc: listItem.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         counter.Qps,
			Qpd:         counter.DayTotal,
			TotalNode: len(ipList),
		}
		outList = append(outList, outItem)
	}
	out := &dto.ServiceListOutput{
		Total: int64(len(outList)),
		List:  outList,
	}
	middleware.ResponseSuccess(c, out)
}

func (service *ServiceController) ServiceAddHttp(c *gin.Context) {
	/*
		1. 参数绑定
		2. ip列表和权重数量校验
		----开事务
			1. 服务是否存在
			2. rule是否存在
			3. 保存
	*/
	params := &dto.ServiceAddHTTPInput{}
	if err := params.BindValidInput(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2001, errors.New("权重与Ip不对应"))
		return
	}
	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	tx, err := db.Begin()
	if err != nil {
		middleware.ResponseError(c, 2001, errors.New("服务忙"))
		return
	}
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	//err.Error() != "[scanner]: empty result"

	if _, err = serviceInfo.Find(c, db); err.Error() != "[scanner]: empty result" {
		tx.Rollback()
		middleware.ResponseError(c, 2002, errors.New("服务已存在"))
		return
	}

	httpUrl := &dao.HttpRule{RuleType: params.RuleType, Rule: params.Rule}
	if _, err := httpUrl.Find(c, db); err.Error() != "[scanner]: empty result" {
		tx.Rollback()
		middleware.ResponseError(c, 2003, errors.New("服务接入前缀或域名已存在"))
		return
	}

	info := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
		LoadType:    public.HttpLoadType,
	}
	if err := info.Insert(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}
	//serviceModel.ID
	httpRule := &dao.HttpRule{
		ServiceID:      info.ID,
		RuleType:       params.RuleType,
		Rule:           params.Rule,
		NeedHttps:      params.NeedHttps,
		NeedStripUri:   params.NeedStripUri,
		NeedWebsocket:  params.NeedWebsocket,
		UrlRewrite:     params.UrlRewrite,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := httpRule.Insert(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientipFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Insert(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	loadbalance := &dao.LoadBalance{
		ServiceID:              info.ID,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  params.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    params.UpstreamIdleTimeout,
		UpstreamMaxIdle:        params.UpstreamMaxIdle,
	}
	if err := loadbalance.Insert(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")

}

func (service *ServiceController) ServiceUpdateHttp(c *gin.Context) {
	params := &dto.ServiceAddHTTPInput{}
	if err := params.BindValidInput(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	// 开事务去更新数据 涉及表： service_info、http_rule、accessController、 loadBalance
	tx, err := db.Begin()
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	info := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
	}
	detail, err := info.ServiceDetail(c, tx)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, err)
		return
	}

	// 更新info表
	detail.ServiceInfo.ServiceDesc = params.ServiceDesc
	detail.ServiceInfo.UpdateAt = time.Now()
	if err = detail.ServiceInfo.Update(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, err)
		return
	}
	detail.HTTPRule.NeedHttps = params.NeedHttps
	detail.HTTPRule.NeedStripUri = params.NeedStripUri
	detail.HTTPRule.NeedWebsocket = params.NeedWebsocket
	detail.HTTPRule.UrlRewrite = params.UrlRewrite
	detail.HTTPRule.HeaderTransfor = params.HeaderTransfor
	if err = detail.HTTPRule.Update(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, err)
		return
	}
	// 更新accessController表
	detail.AccessControl.OpenAuth = params.OpenAuth
	detail.AccessControl.BlackList = params.BlackList
	detail.AccessControl.WhiteList = params.WhiteList
	detail.AccessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	detail.AccessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err = detail.AccessControl.Update(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, err)
		return
	}

	loadbalance := detail.LoadBalance
	loadbalance.RoundType = params.RoundType
	loadbalance.IpList = params.IpList
	loadbalance.WeightList = params.WeightList
	loadbalance.UpstreamConnectTimeout = params.UpstreamConnectTimeout
	loadbalance.UpstreamHeaderTimeout = params.UpstreamHeaderTimeout
	loadbalance.UpstreamIdleTimeout = params.UpstreamIdleTimeout
	loadbalance.UpstreamMaxIdle = params.UpstreamMaxIdle
	if err := loadbalance.Update(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	// 发消息 更新缓存
	redisDB, err := lib.RedisConnFactory("default")
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2003, err)
		return
	}
	if err = detail.UpdateGlobalService(c, redisDB); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2004, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")

}

func (service *ServiceController) ServiceDetail(c *gin.Context) {

	param := &dto.ServiceDetailInput{}
	if err := param.BindValidInput(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	info := &dao.ServiceInfo{ID: param.ServiceId}
	detail, err := info.ServiceDetail(c, db)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, detail)

}

func (service *ServiceController) ServiceDelete(c *gin.Context) {
	param := &dto.ServiceDeleteInput{}
	if err := param.BindValidInput(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	info := &dao.ServiceInfo{ID: param.ServiceId}
	err = info.Delete(c, db)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
}

func (service *ServiceController) ServiceAddTcp(c *gin.Context) {
	param := &dto.ServiceAddTcpInput{}
	if err := param.BindValidInput(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	// 校验ip和权重数量一致
	// 校验端口是否占用

	if len(strings.Split(param.WeightList, ",")) != len(strings.Split(param.IpList, ",")) {
		middleware.ResponseError(c, 2002, errors.New("ip和权重数量不匹配"))
		return
	}

	info := &dao.ServiceInfo{ServiceName: param.ServiceName}
	_, err = info.Find(c, db)
	if err == nil {
		middleware.ResponseError(c, 2002, errors.New("服务名被占用"))
		return
	}

	grpc := &dao.GrpcRule{Port: param.Port}
	_, err = grpc.Find(c, db)
	if err == nil {
		middleware.ResponseError(c, 2002, errors.New("端口被占用"))
		return
	}
	tcp := &dao.TcpRule{Port: param.Port}
	_, err = tcp.Find(c, db)
	if err == nil {
		middleware.ResponseError(c, 2002, errors.New("端口被占用"))
		return
	}
	// 开事务保存。有三张表：info 、load_balance、tcpRule表
	tx, err := db.Begin()
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		tx.Rollback()
		return
	}
	info = &dao.ServiceInfo{
		LoadType:    public.TcpLoadType,
		ServiceName: param.ServiceName,
		ServiceDesc: param.ServiceDesc,
		IsDelete:    public.NoDelete,
	}
	err = info.Insert(c, tx)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		tx.Rollback()
		return
	}

	balance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  param.RoundType,
		IpList:     param.IpList,
		WeightList: param.WeightList,
		ForbidList: param.BlackList,
	}
	err = balance.Insert(c, tx)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		tx.Rollback()
		return
	}

	tcp = &dao.TcpRule{
		ServiceID: info.ID,
		Port:      param.Port,
	}
	err = tcp.Insert(c, tx)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		tx.Rollback()
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          param.OpenAuth,
		BlackList:         param.BlackList,
		WhiteList:         param.WhiteList,
		ClientIPFlowLimit: param.ClientIpFlowLimit,
		ServiceFlowLimit:  param.ServiceFlowLimit,
	}
	if err := accessControl.Insert(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	if err = tx.Commit(); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")

}

func (service *ServiceController) ServiceUpdateTcp(c *gin.Context) {
	params := &dto.ServiceUpdateTcpInput{}
	if err := params.BindValidInput(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	// 开事务去更新数据 涉及表： service_info、http_rule、accessController、 loadBalance
	tx, err := db.Begin()
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	info := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
	}
	detail, err := info.ServiceDetail(c, tx)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, err)
		return
	}

	// 更新info表
	detail.ServiceInfo.ServiceDesc = params.ServiceDesc
	detail.ServiceInfo.UpdateAt = time.Now()
	if err = detail.ServiceInfo.Update(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, err)
		return
	}

	// 更新accessController表
	detail.AccessControl.OpenAuth = params.OpenAuth
	detail.AccessControl.BlackList = params.BlackList
	detail.AccessControl.WhiteList = params.WhiteList
	detail.AccessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	detail.AccessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err = detail.AccessControl.Update(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, err)
		return
	}

	loadbalance := detail.LoadBalance
	loadbalance.RoundType = params.RoundType
	loadbalance.IpList = params.IpList
	loadbalance.WeightList = params.WeightList
	if err := loadbalance.Update(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	// todo update tcp发消息 更新缓存
	/*	redisDB, err := lib.RedisConnFactory("default")
		if err!=nil {
			tx.Rollback()
			middleware.ResponseError(c, 2003, err)
			return
		}
		if err = detail.UpdateGlobalService(c, redisDB);err!=nil{
			tx.Rollback()
			middleware.ResponseError(c, 2004, err)
			return
		}*/
	tx.Commit()
	middleware.ResponseSuccess(c, "")
}

func (service *ServiceController) ServiceAddGrpc(c *gin.Context) {
	param := &dto.ServiceAddGrpcInput{}
	if err := param.BindValidInput(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	// 校验ip和权重数量一致
	// 校验端口是否占用

	if len(strings.Split(param.WeightList, ",")) != len(strings.Split(param.IpList, ",")) {
		middleware.ResponseError(c, 2002, errors.New("ip和权重数量不匹配"))
		return
	}

	info := &dao.ServiceInfo{ServiceName: param.ServiceName}
	_, err = info.Find(c, db)
	if err == nil {
		middleware.ResponseError(c, 2002, errors.New("服务名被占用"))
		return
	}

	grpc := &dao.GrpcRule{Port: param.Port}
	_, err = grpc.Find(c, db)
	if err == nil {
		middleware.ResponseError(c, 2002, errors.New("端口被占用"))
		return
	}
	tcp := &dao.TcpRule{Port: param.Port}
	_, err = tcp.Find(c, db)
	if err == nil {
		middleware.ResponseError(c, 2002, errors.New("端口被占用"))
		return
	}

	// 开事务保存 三张表：info、grpc、load_balance
	tx, err := db.Begin()
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	info.ServiceDesc = param.ServiceDesc
	info.LoadType = public.GrpcLoadType
	info.IsDelete = public.NotNeedHttps
	if err = info.Insert(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, err)
		return
	}

	grpc.ServiceID = info.ID
	grpc.Port = param.Port
	grpc.HeaderTransfor = param.HeaderTransfor
	if err = grpc.Insert(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, err)
		return
	}

	balance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  param.RoundType,
		IpList:     param.IpList,
		WeightList: param.WeightList,
		ForbidList: param.BlackList,
	}
	err = balance.Insert(c, tx)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		tx.Rollback()
		return
	}

	if err = tx.Commit(); err != nil {
		middleware.ResponseError(c, 2003, err)
		tx.Rollback()
		return
	}

	middleware.ResponseSuccess(c, "")

}

func (service *ServiceController) ServiceUpdateGrpc(c *gin.Context){
	params := &dto.ServiceUpdateTcpInput{}
	if err := params.BindValidInput(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	// 开事务去更新数据 涉及表： service_info、http_rule、accessController、 loadBalance
	tx, err := db.Begin()
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	info := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
	}
	detail, err := info.ServiceDetail(c, tx)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, err)
		return
	}

	// 更新info表
	detail.ServiceInfo.ServiceDesc = params.ServiceDesc
	detail.ServiceInfo.UpdateAt = time.Now()
	if err = detail.ServiceInfo.Update(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, err)
		return
	}

	// 更新accessController表
	detail.AccessControl.OpenAuth = params.OpenAuth
	detail.AccessControl.BlackList = params.BlackList
	detail.AccessControl.WhiteList = params.WhiteList
	detail.AccessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	detail.AccessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err = detail.AccessControl.Update(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, err)
		return
	}

	loadbalance := detail.LoadBalance
	loadbalance.RoundType = params.RoundType
	loadbalance.IpList = params.IpList
	loadbalance.WeightList = params.WeightList
	if err := loadbalance.Update(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	// todo update tcp发消息 更新缓存
	/*	redisDB, err := lib.RedisConnFactory("default")
		if err!=nil {
			tx.Rollback()
			middleware.ResponseError(c, 2003, err)
			return
		}
		if err = detail.UpdateGlobalService(c, redisDB);err!=nil{
			tx.Rollback()
			middleware.ResponseError(c, 2004, err)
			return
		}*/
	tx.Commit()
	middleware.ResponseSuccess(c, "")
}