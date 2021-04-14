package public

import (
	"golang.org/x/time/rate"
)

type FlowLimiter struct{
	FlowLimitMap map[string]*rate.Limiter
}
var FLowLimitHandler *FlowLimiter

func init(){
	FLowLimitHandler = &FlowLimiter{
		FlowLimitMap: make(map[string]*rate.Limiter),
	}
}

func (r *FlowLimiter)GetRateLimiter(serviceName string, qps int)*rate.Limiter{
	if limiter, ok := r.FlowLimitMap[serviceName];ok {
		return limiter
	}
	r.FlowLimitMap[serviceName] = rate.NewLimiter(rate.Limit(qps), qps*3)
	return r.FlowLimitMap[serviceName]
}