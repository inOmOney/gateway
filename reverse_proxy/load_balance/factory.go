package load_balance

type LBType int

const (
	RandomType LBType = iota
	RoundRobinType
	WeightRoundRobinType
	IpHashType
)

func LBFactory(lbType LBType, conf *LoadBalanceConfig) LoadBalance {
	switch lbType {
	case RandomType:
		rb := &RandomBalance{}
		rb.SetConfig(conf)
		conf.ob = rb
		rb.Update()
		return rb
	//todo 各种负载均衡策略
	case RoundRobinType:
		rr := &RoundRobin{}
		rr.SetConfig(conf)
		conf.ob = rr
		rr.Update()
		return rr
	case WeightRoundRobinType:
		wlb := &WeightLoadBalance{}
		wlb.SetConf(conf)
		conf.ob = wlb
		wlb.Update()
		return wlb
	case IpHashType:
		ch := NewConsistentHash(conf)
		conf.ob = ch
		ch.Update()
		return ch
	default:
		rb := &RandomBalance{}
		rb.SetConfig(conf)
		rb.Update()
		return rb
	}
}
