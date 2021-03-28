package load_balance

import "errors"

type RoundRobin struct{
	curIndex int
	addrs []string
	conf *LoadBalanceConfig
}

func (rr *RoundRobin) Add(addr... string) error{
	if len(addr) == 0  {
		return errors.New("at least 1 addr")
	}
	rr.addrs = append(rr.addrs, addr...)
	return nil
}

func (rr *RoundRobin) Get(p string)string{
	if rr.curIndex >= len(rr.addrs) {
		rr.curIndex = 0
	}
	curAddr := rr.addrs[rr.curIndex]
	rr.curIndex++
	return curAddr
}
func (rr *RoundRobin)Update(){
	rr.addrs = rr.conf.Alive
}

func (rr *RoundRobin)SetConfig(config *LoadBalanceConfig){
	rr.conf = config
}


