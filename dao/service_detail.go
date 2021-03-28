package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"gateway/dto"
	"gateway/lib"
	"gateway/log"
	"gateway/public"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"net/http/httptest"
	"strings"
	"sync"
	"time"
)

type ServiceDetail struct {
	ServiceInfo   *ServiceInfo   `json:"info" description:"基本信息"`
	HTTPRule      *HttpRule      `json:"http_rule" description:"http_rule"`
	TCPRule       *TcpRule       `json:"tcp_rule" description:"tcp_rule"`
	GRPCRule      *GrpcRule      `json:"grpc_rule" description:"grpc_rule"`
	LoadBalance   *LoadBalance   `json:"load_balance" description:"load_balance"`
	AccessControl *AccessControl `json:"access_control" description:"access_control"`
}

var SvcManager *ServiceManager

func init() {
	SvcManager = ServiceManagerNew()
}

type ServiceManager struct {
	ServiceMap map[string]*ServiceDetail
	Lock       sync.RWMutex
	init       sync.Once
}

func ServiceManagerNew() *ServiceManager {
	return &ServiceManager{
		ServiceMap: map[string]*ServiceDetail{},
		Lock:       sync.RWMutex{},
		init:       sync.Once{},
	}
}

func (s *ServiceManager) LoadOnceService() {
	s.init.Do(func() {
		tempContext, _ := gin.CreateTestContext(httptest.NewRecorder())
		db, err := lib.GetDBPool("default")
		if err != nil {
			log.Fatal("Service加载错误")
		}

		info := &ServiceInfo{}
		param := &dto.ServiceListInput{Info: "test", PageSize: 9999, PageNo: 1}
		serviceInfos, err := info.ServiceInfoPage(tempContext, db, param)
		if err != nil {
			log.Fatal("Service加载错误")
		}

		for _, temp := range serviceInfos { // 通过serviceInfo拿到每个info的各种参数Detail
			serviceInfo := temp
			detail, err := serviceInfo.ServiceDetail(tempContext, db)
			if err != nil {
				log.Fatal("Service加载错误")
			}
			s.ServiceMap[serviceInfo.ServiceName] = detail
		}

		go func() {
			c, _ := lib.RedisConnFactory("default")
			psc := redis.PubSubConn{Conn: c}

			psc.PSubscribe("update_service_*")
			for {
				switch v := psc.Receive().(type) {
				case redis.Message:
					{
						var msg ServiceDetailMsg
						if err := json.Unmarshal(v.Data, &msg); err != nil {
							fmt.Println("反序列化出错", err)
						}
						s.ServiceMap[msg.ServiceName] = &msg.ServiceDetail

						delete(LoadBalancerHandler.LoadBalanceMap, msg.ServiceName)// 清空LoadBalance缓存
						fmt.Println("更新Detail:", msg.ServiceName)
					}
				case redis.Subscription:
					fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
				case error:
					fmt.Printf("接收消息出错", err)
				}
				time.Sleep(3 * time.Second)
			}

		}()

	})
}

// detailManager 对应请求的类型来获取serviceDetail
func (s *ServiceManager) GetDetailFromReq(c *gin.Context) (*ServiceDetail, error) {
	host := c.Request.Host
	path := c.Request.URL.Path

	for _, detail := range s.ServiceMap {
		if detail.ServiceInfo.LoadType != public.HttpLoadType {
			continue
		}
		if detail.HTTPRule.Rule == host || strings.HasPrefix(detail.HTTPRule.Rule, path) {
			return detail, nil
		}
	}
	return nil, errors.New("没有找到相应的服务")

}

type ServiceDetailMsg struct {
	ServiceName   string        `json:"service_name"`
	ServiceDetail ServiceDetail `json:"service_detail"`
}

func (s *ServiceDetail) UpdateGlobalService(c *gin.Context, db redis.Conn) error { // 发送二进制
	msg := ServiceDetailMsg{
		ServiceName:   s.ServiceInfo.ServiceName,
		ServiceDetail: *s,
	}
	marshal, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = lib.RedisLogDo(public.GetGinTraceContext(c), db, "PUBLISH", lib.GetStringConf("proxy.update.channel")+s.ServiceInfo.ServiceName, marshal)
	fmt.Println("发送消息成功")
	return err
}

