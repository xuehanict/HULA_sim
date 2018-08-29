package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	UpdateWindow = 1
	SendCycle    = 5
	BufferSize   = 10
)

type RouterID int

type Probe struct {
	dest    RouterID
	pathDis int
	pathBls int64
}

type HulaRouter struct {
	// RouterID
	ID RouterID

	// ProbeUpdateTable 记录dest的最近更新时间
	ProbeUpdateTable map[RouterID]int64

	// BestHopsTable 记录下一跳路由器ID
	BestHopsTable map[RouterID]HopTableEntry

	Neighbors []RouterID

	LinkBase map[string]*HulaLink

	RouterBase map[RouterID]*HulaRouter

	// MessagePool 作为信息池
	MessagePool chan *Probe

	// 计时器定期执行任务
	timer *time.Ticker

	wg sync.WaitGroup

	quit chan struct{}
}

type HopTableEntry struct {
	nextHop  RouterID
	capacity int64
}

type HulaLink struct {
	// R1 R2 分别指示两端路由器
	R1 RouterID
	R2 RouterID
	// capacity 模拟通道容量
	capacity int64
}

func (r *HulaRouter) start() {
	r.wg.Add(1)
	defer r.wg.Done()
	go func() {
		for {
			select {
			case probe := <-r.MessagePool:
				r.handleProbe(probe)
			case <-r.quit:
				return
			case <-r.timer.C:
				for _, neighbor := range r.Neighbors {
					link := r.getLink(neighbor)
					if link == nil {
						fmt.Printf("router %v can't find the "+
							"link between %v", r.ID, neighbor)
						continue
					}
					probe := r.createProbe(link.capacity)
					err := r.sendProbeToRouter(neighbor, probe)
					if err != nil {
						fmt.Printf("router %v send probe to "+
							"%v failed : %v", r.ID, neighbor, err)
						continue
					}
				}
			}
		}
	}()
}

func (r *HulaRouter) stop() {
	close(r.quit)
}

func (r *HulaRouter) createProbe(bls int64) *Probe {
	probe := &Probe{
		dest:    r.ID,
		pathDis: 0,
		pathBls: bls,
	}
	return probe
}

func (r *HulaRouter) handleProbe(p *Probe) error {


	return nil
}

func (r *HulaRouter) printBestHopTable() {
	for entry := range r.BestHopsTable {
		fmt.Printf("%v\n", entry)
	}
}

func (r *HulaRouter) sendProbeToRouter(id RouterID, probe *Probe) error {
	neighbor, ok := r.RouterBase[id]
	if ok != true {
		return fmt.Errorf("can't find the router id : %v", id)
	}
	neighbor.MessagePool <- probe
	return nil
}

func newRouter(id RouterID) *HulaRouter {
	router := &HulaRouter{
		ID:id,
		ProbeUpdateTable: make(map[RouterID]int64),
		BestHopsTable: make(map[RouterID]HopTableEntry),
		Neighbors: make([]RouterID,0),
		MessagePool: make(chan *Probe, BufferSize),
		timer: time.NewTicker(SendCycle * time.Second),
		wg: sync.WaitGroup{},
		quit: make(chan struct{}),
	}
	return router
}

func (r *HulaRouter) getLink(neighbor RouterID) *HulaLink {
	var link *HulaLink
	link, ok := r.LinkBase[newLinkKey(r.ID, neighbor)]
	if ok == true {
		return link
	}
	link, ok = r.LinkBase[newLinkKey(neighbor, r.ID)]
	if ok == true {
		return link
	}
	return nil
}

// addLink adds a link between two
func addLink(r1, r2 RouterID, capacity int64, linkBase map[string]*HulaLink) error {
	linkKey1 := newLinkKey(r1, r2)
	linkKey2 := newLinkKey(r2, r1)
	link := &HulaLink{
		R1:       r1,
		R2:       r2,
		capacity: capacity,
	}
	_, ok1 := linkBase[linkKey1]
	_, ok2 := linkBase[linkKey2]
	ok := ok1 || ok2
	if ok {
		return fmt.Errorf("link: %v <-----> %v exsist", r1, r2)
	}
	linkBase[linkKey1] = link
	return nil
}
