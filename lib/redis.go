package lib

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"math/rand"
	"time"
)

func RedisConnFactory(name string) (redis.Conn, error) {
	if ConfRedisMap != nil && ConfRedisMap.List != nil {
		for confName, cfg := range ConfRedisMap.List {
			if name == confName {
				randHost := cfg.ProxyList[rand.Intn(len(cfg.ProxyList))]
				if cfg.ConnTimeout == 0 {
					cfg.ConnTimeout = 50
				}
				if cfg.ReadTimeout == 0 {
					cfg.ReadTimeout = 100
				}
				if cfg.WriteTimeout == 0 {
					cfg.WriteTimeout = 100
				}
				c, err := redis.Dial(
					"tcp",
					randHost,
				)
				if err != nil {
					return nil, err
				}
				if cfg.Password != "" {
					if _, err := c.Do("AUTH", cfg.Password); err != nil {
						c.Close()
						return nil, err
					}
				}
				if cfg.Db != 0 {
					if _, err := c.Do("SELECT", cfg.Db); err != nil {
						c.Close()
						return nil, err
					}
				}
				return c, nil
			}
		}
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