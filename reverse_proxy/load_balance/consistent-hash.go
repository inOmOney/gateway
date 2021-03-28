package load_balance

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
)

type Point []uint32

func (p Point) Len() int {
	return len(p)
}
func (p Point) Less(i, j int) bool {
	return p[i] <= p[j]
}
func (p Point) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type ConsistentHash struct {
	points     Point
	point2addr map[uint32]string

	conf *LoadBalanceConfig
}

func NewConsistentHash(config *LoadBalanceConfig) *ConsistentHash {
	return &ConsistentHash{
		points:     []uint32{},
		point2addr: map[uint32]string{},
		conf : config,
	}
}

func (ch *ConsistentHash)Update(){
	newCH := ConsistentHash{
		points:     []uint32{},
		point2addr: map[uint32]string{},
		conf:       ch.conf,
	}
	newCH.Add(ch.conf.Alive...)
	*ch = newCH
}

func (ch *ConsistentHash) Add(param ...string)error {
	if len(param) == 0 {
		return errors.New("param必须大于0个")
	}
	virtualPoint := 10
	for i := 0; i < len(param); i++ {
		for j := 0; j < virtualPoint; j++ { // 为每个节点生成10个虚拟节点
			hash := crc32.ChecksumIEEE([]byte( strconv.Itoa(j) + param[i] ))
			ch.points = append(ch.points, hash)
			ch.point2addr[hash] = param[i]
		}
		sort.Sort(ch.points)
	}
	return nil
}

func (ch *ConsistentHash) Get(clientAddr string) string {
	hash := crc32.ChecksumIEEE([]byte(clientAddr))

	index := sort.Search(len(ch.points), func(i int) bool {
		return ch.points[i] >= hash
	}) // 找到第一个大于请求hash结果的节点

	if index == len(ch.points) {
		index = 0
	}

	return ch.point2addr[ch.points[index]]
}

