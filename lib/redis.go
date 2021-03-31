package lib

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

func RedisConnFactory(name string) (redis.Conn, error) {
	value, ok:= RedisPoolMap[name]
	if ok {
		return value.Get(),nil
	}
	return nil, errors.New("create redis conn fail")
}

func RedisLogDo(trace *TraceContext, c redis.Conn, commandName string, args ...interface{}) (interface{}, error) {
	startExecTime := time.Now()
	reply, err := c.Do(commandName, args...)
	endExecTime := time.Now()
	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method":    commandName,
			"err":       err,
			"bind":      args,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	} else {
		replyStr, _ := redis.String(reply, nil)
		Log.TagInfo(trace, "_com_redis_success", map[string]interface{}{
			"method":    commandName,
			"bind":      args,
			"reply":     replyStr,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	}
	return reply, err
}

func RedisLogPipelining(trace *TraceContext, c redis.Conn, f func(conn redis.Conn) )(interface{}, error) {


	startExecTime := time.Now()
	f(c)
	c.Flush()
	reply, err := c.Receive()
	endExecTime := time.Now()

	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method":    "Update Flow",
			"err":       err,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	} else {
		replyStr, _ := redis.String(reply, nil)
		Log.TagInfo(trace, "_com_redis_success", map[string]interface{}{
			"method":  "Update Flow",
			"reply":     replyStr,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	}
	return reply, err
}

//通过配置 执行redis
func RedisConfDo(trace *TraceContext, name string, commandName string, args ...interface{}) (interface{}, error) {
	c, err := RedisConnFactory(name)
	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method": commandName,
			"err":    errors.New("RedisConnFactory_error:" + name),
			"bind":   args,
		})
		return nil, err
	}
	defer c.Close()

	startExecTime := time.Now()
	reply, err := c.Do(commandName, args...)
	endExecTime := time.Now()
	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method":    commandName,
			"err":       err,
			"bind":      args,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	} else {
		replyStr, _ := redis.String(reply, nil)
		Log.TagInfo(trace, "_com_redis_success", map[string]interface{}{
			"method":    commandName,
			"bind":      args,
			"reply":     replyStr,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	}
	return reply, err
}

func RedisLogSub(trace *TraceContext, c redis.Conn, subName string, msgChannel chan []byte) (interface{}, error) {
	psc := redis.PubSubConn{Conn: c}
	psc.Subscribe(subName)
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
			Log.TagInfo(trace, "_com_redis_sub_success", map[string]interface{}{
				"channel_name":    v.Channel,
				"data":      v.Data,
				"time": fmt.Sprintf("%fs", time.Now()),
			})
			msgChannel <- v.Data
		case redis.Subscription:
		case error:
			return nil, v
		}

	}
}
