package load_balance

import (
	"errors"
	"math/rand"
	"time"
)

type RandomBalance struct{
	addr []string
	conf *LoadBalanceConfig
}
func init(){
	rand.Seed(time.Now().Unix())
}
func (rb *RandomBalance)Add(addr ...string) error{
	if len(addr) == 0 {
		return errors.New("at least 1 addr")
	}
	rb.addr = append(rb.addr, addr...)
	return nil
}

func (rb *RandomBalance)Get(p string)string{
	index := rand.Intn(len(rb.addr))
	return rb.addr[index]
}

func (rb *RandomBalance)Update(){
	rb.addr = rb.conf.Alive
}

func (rb *RandomBalance)SetConfig(config *LoadBalanceConfig){
	rb.conf = config
}