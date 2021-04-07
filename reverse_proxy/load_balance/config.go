package load_balance

import (
	"encoding/json"
	"fmt"
	"gateway/lib"
	"gateway/public"
	"github.com/gomodule/redigo/redis"
	"net"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type LoadBalanceConfig struct { // 都加上了协议
	Alive        []string
	IpWeightList map[string]int
	ob           Observer
}

// ipWeight:  key:127.0.0.1:8080 value:10
func NewConfig(schema, serviceName, ip, weight string) *LoadBalanceConfig {

	rawIp := strings.Split(ip, ",")
	rawWeight := strings.Split(weight, ",")

	ipWeightMap := make(map[string]int)
	for i := 0; i < len(rawIp); i++ {
		weight, _ := strconv.Atoi(rawWeight[i])
		ipWeightMap[schema+rawIp[i]] = weight
	}

	//ip加上协议
	tempIpList := []string{}
	for _,key := range rawIp {
		tempIpList = append(tempIpList, fmt.Sprintf("%s%s", schema, key))
	}
	config := &LoadBalanceConfig{
		Alive:        tempIpList,
		IpWeightList: ipWeightMap,
	}
	config.WatchAlive()
	config.WatchChange(serviceName)
	return config
}

// aliveIp格式: ip1,ip2,ip3
// 返回格式 ip,weight,ip,weight
func (config *LoadBalanceConfig) UseIpGetIpWeight() []string {
	res := []string{}
	for _, ip := range config.Alive {
		res = append(res, ip, strconv.Itoa(config.IpWeightList[ip]))
	}
	return res
}

func (config *LoadBalanceConfig) WatchChange(serviceName string) { // detail变化 同步更新iplist

	go func() {
		c, _ := lib.RedisConnFactory("default")
		psc := redis.PubSubConn{Conn: c}

		psc.Subscribe("update_service_" + serviceName)
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				{
					var data ServiceDetailMsg
					if err := json.Unmarshal(v.Data, &data); err != nil {
						fmt.Println("反序列化出错", err)
					}
					rawIp := strings.Split(data.ServiceDetail.LoadBalance.IpList, ",")
					rawWeight := strings.Split(data.ServiceDetail.LoadBalance.WeightList, ",")
					ipWeightMap := make(map[string]int)

					schema := ""
					if data.ServiceDetail.ServiceInfo.LoadType == public.HttpLoadType {
						schema = "http://"
					}
					for i := 0; i < len(rawIp); i++ {
						weight, _ := strconv.Atoi(rawWeight[i])
						ipWeightMap[schema+rawIp[i]] = weight
					}
					config.IpWeightList = ipWeightMap
					fmt.Println("更新LoadBalance中IpList")
				}
			case redis.Subscription:
				fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
			case error:
				fmt.Printf("接收消息出错%s", v.Error())
			}
		}
	}()
}

type ServiceDetailMsg struct {
	ServiceName   string `json:"service_name"`
	ServiceDetail struct {
		ServiceInfo struct {
			LoadType int `json:"load_type" description:"负载类型 0=http 1=tcp 2=grpc"`
		}
		LoadBalance struct {
			IpList     string `json:"ip_list" gorm:"column:ip_list" description:"ip列表"`
			WeightList string `json:"weight_list" gorm:"column:weight_list" description:"权重列表"`
		} `json:"load_balance" description:"load_balance"`
	} `json:"service_detail"`
}

func (config *LoadBalanceConfig) WatchAlive() { // 定时探测服务
	//todo 探测失效服务
	go func() {
		isAlive := make(map[string]int) // 记录ip探测失败的次数
		for _, temp := range config.Alive {
			isAlive[temp] = 0
		}
		for {
			aliveIp := []string{}
			for ip := range config.IpWeightList {
				conn, err := net.DialTimeout("tcp", strings.Split(ip, "//")[1], 2*time.Second)
				if err == nil {
					conn.Close()
					isAlive[ip] = 0
				} else {
					isAlive[ip]++
				}
				if isAlive[ip] < 2 {
					aliveIp = append(aliveIp, ip)
				}
			}
			sort.Strings(aliveIp)
			sort.Strings(config.Alive)

			if !reflect.DeepEqual(aliveIp, config.Alive) {
				config.Alive = aliveIp
				fmt.Printf("服务列表更新: %s\n", config.Alive)
				config.ob.Update()
			}
			fmt.Printf("服务探针检测 各项服务正常")
			time.Sleep(5 * time.Second)
		}
	}()
}
