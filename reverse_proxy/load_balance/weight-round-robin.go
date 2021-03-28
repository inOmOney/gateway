package load_balance

import (
	"errors"
	"strconv"
)

type WeightLoadBalance struct {
	curIndex int
	addrs    []*WeightNode

	conf *LoadBalanceConfig
}

type WeightNode struct {
	addr string

	weight int

	currentWeight int
}

// addr格式 ip,weight,ip,weight
func (lb *WeightLoadBalance) Add(addr ...string) error {
	temp := []*WeightNode{}
	if len(addr)%2 != 0 {
		return errors.New("ip与权重不匹配")
	}
	for i := 0; i < len(addr); i = i + 2 {
		wNum, err := strconv.Atoi(addr[i+1]) // 权重
		if err != nil {
			return errors.New("权重需为数字")
		}
		node := &WeightNode{
			addr:   addr[i],
			weight: wNum,
		}
		temp = append(temp, node)
	}
	lb.addrs = temp
	return nil
}

func (lb *WeightLoadBalance) Get(p string) string {
	total := 0 // 所有节点的权重之和，用来周期按权重轮训
	var currNode *WeightNode
	for i := 0; i < len(lb.addrs); i++ {
		total += lb.addrs[i].weight
		lb.addrs[i].currentWeight += lb.addrs[i].weight
		if currNode == nil || currNode.currentWeight < lb.addrs[i].currentWeight {
			currNode = lb.addrs[i]
		}
	}
	currNode.currentWeight -= total
	return currNode.addr
}

func (lb *WeightLoadBalance) SetConf(lbConf *LoadBalanceConfig)  {
	lb.conf = lbConf
}

// 运行本方法时 说明lb.conf.Alive 已经是最新的可用ip。格式：http://127.0.0.1:8888
func (lb *WeightLoadBalance) Update()  {
	weight := lb.conf.UseIpGetIpWeight()
	lb.Add(weight...)
}
