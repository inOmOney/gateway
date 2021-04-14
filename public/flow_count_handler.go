package public

import (
	"gateway/lib"
	"gateway/log"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"sync/atomic"
	"time"
)

type FlowCountManager struct {
	FlowCountMap map[string]*FlowCountInfo
}

type FlowCountInfo struct {
	ServiceName string //全局的ServiceName是Global
	TimeFlow    int64  //统计周期内的访问
	UnixTime    int64

	DayTotal int64
	Qps      int64
}

var FlowCountHandler *FlowCountManager

func init() {
	FlowCountHandler = &FlowCountManager{
		FlowCountMap: map[string]*FlowCountInfo{},
	}
}

//serviceName 可能为全局统计的常量 或者为 app租户的ID
func (o *FlowCountManager) GetServiceCountHandler(serviceName string, c *gin.Context) *FlowCountInfo {
	countHandler, ok := o.FlowCountMap[serviceName]
	rdb, _ := lib.RedisConnFactory("default")

	if ok {
		return countHandler
	} else {
		o.FlowCountMap[serviceName] = &FlowCountInfo{
			UnixTime:    time.Now().Unix(),
			ServiceName: serviceName,
		}
		countHandler = o.FlowCountMap[serviceName]
		DayKey := countHandler.GetDayKey()
		total, _ := redis.Int64(lib.RedisLogDo(GetGinTraceContext(c), rdb, "GET", DayKey))
		countHandler.DayTotal = total
		countHandler.UnixTime = time.Now().Unix()
	}
	go func() {
		timer := time.NewTicker(Interval)
		for true {
			<-timer.C
			// 增加redis中对应时间段的统计值 1.当天维度 2.时段维度
			flow := countHandler.TimeFlow
			countHandler.TimeFlow = 0

			DayKey := countHandler.GetDayKey()
			HourKey := countHandler.GetHourKey()

			//dashboard模式运行下面代码会无视
			lib.RedisLogPipelining(GetGinTraceContext(c), rdb, func(conn redis.Conn) {
				conn.Send("INCRBY", DayKey, flow)
				conn.Send("EXPIRE", DayKey, 60*60*12*2)
				conn.Send("INCRBY", HourKey, flow)
				conn.Send("EXPIRE", HourKey, 60*60*12*2)
			},
			)

			//1. 更新qps
			//2. 更新总访问数量
			//3. 更新时间
			total, err := redis.Int64(lib.RedisLogDo(GetGinTraceContext(c), rdb, "GET", DayKey))
			if err != nil {
				log.Info("reqCounter.GetDayData err", err)
				continue
			}
			now := time.Now().Unix()
			if now == countHandler.UnixTime {
				continue
			} else {
				countHandler.Qps = (total - countHandler.DayTotal) / (now - countHandler.UnixTime)
			}
			countHandler.DayTotal = total
			countHandler.UnixTime = now
		}
	}()
	return o.FlowCountMap[serviceName]
}

func (o *FlowCountInfo) Increase() {
	atomic.AddInt64(&o.TimeFlow, 1)
}

//FlowCountDay_28_[ServiceName]
func (o *FlowCountInfo) GetDayKey() string {
	dayStr := strconv.Itoa(time.Now().Day())
	return FlowCountDayServicePrefix + dayStr + "_" + o.ServiceName
}

//FlowCountHour_28_00_[ServiceName]
func (o *FlowCountInfo) GetHourKey() string {
	hourStr := time.Now().Format("0215")
	return FlowCountHourServicePrefix + hourStr + "_" + o.ServiceName
}

func GetTodayFlow(serviceName string, redisConn redis.Conn, c *gin.Context) ([]int64, error) {
	dayNum := time.Now().Day()
	dayStr := ""
	if dayNum < 10{
		dayStr = "0"+ strconv.Itoa(dayNum)
	}else{
		dayStr = strconv.Itoa(dayNum)
	}
	hour := time.Now().Hour()

	var result []int64
	for i := 0; i <= hour; i++ {
		key := ""
		if i < 10 {
			key = FlowCountHourServicePrefix + dayStr + "0" + strconv.Itoa(i) + "_" + serviceName
		} else {
			key = FlowCountHourServicePrefix + dayStr + strconv.Itoa(i) + "_" + serviceName
		}
		timeFlow, err := redis.Int64(lib.RedisLogDo(GetGinTraceContext(c), redisConn, "GET", key))
		if err != nil && err.Error() != "redigo: nil returned" {
			return nil, err
		}
		result = append(result, timeFlow)
	}
	return result, nil
}
func GetYesterdayFlow(serviceName string, redisConn redis.Conn, c *gin.Context) ([]int64, error) {
	dayNum := time.Now().Day() - 1
	var result []int64
	for i := 0; i < 24; i++ {
		key := ""
		if i < 10 {
			key = FlowCountHourServicePrefix + strconv.Itoa(dayNum) + "0" + strconv.Itoa(i) + "_" + serviceName
		} else {
			key = FlowCountHourServicePrefix + strconv.Itoa(dayNum) + strconv.Itoa(i) + "_" + serviceName
		}
		timeFlow, err := redis.Int64(lib.RedisLogDo(GetGinTraceContext(c), redisConn, "GET", key))
		if err != nil && err.Error() != "redigo: nil returned" {
			return nil, err
		}
		result = append(result, timeFlow)
	}
	return result, nil
}
