package main

import (
	"fmt"
	"sync"
	"time"
	"os"
)

const (
	UpdateWindow = 1
	SendCycle    = 1
	BufferSize   = 1000
)

type RouterID int

type Probe struct {
	dest    RouterID
	upper   RouterID
	pathDis int
	pathBls int64
}

type HulaRouter struct {
	// RouterID
	ID RouterID

	// ProbeUpdateTable 记录dest的最近更新时间
	ProbeUpdateTable map[RouterID]int64

	// BestHopsTable 记录下一跳路由器ID
	BestHopsTable map[RouterID]*HopTableEntry

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
	upperHop RouterID
	dis      int
	capacity int64
	updated  bool
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
					probe := r.newProbe(link.capacity)
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

func (r *HulaRouter) newProbe(bls int64) *Probe {
	probe := &Probe{
		dest:    r.ID,
		upper:   r.ID,
		pathDis: 0,
		pathBls: bls,
	}
	return probe
}

func (r *HulaRouter) handleProbe(p *Probe) error {

	if p.dest == r.ID {
		return  nil
	}
	if p.dest == 5 && p.upper == 9 && r.ID == 10 {
		fmt.Printf("10 收到9 发来的5的probe:%v\n", p)
		r.printBestHopTable()
	}
	bestHopEntry, ok := r.BestHopsTable[p.dest]
	// 如果没有在路由表中，那么我们将其添加到路由表中,并且将这个probe
	// 迅速发给它的所有邻居。
	if !ok {
		r.BestHopsTable[p.dest] = &HopTableEntry{
			upperHop: p.upper,
			capacity: p.pathBls,
			dis:      p.pathDis + 1,
		}
		r.ProbeUpdateTable[p.dest] = time.Now().Unix()
		for _, neighbor := range r.Neighbors {
			probe := &Probe{
				dest:    p.dest,
				upper:   r.ID,
				pathDis: p.pathDis + 1,
			}
			if p.pathBls > r.getLink(neighbor).capacity {
				probe.pathBls = r.getLink(neighbor).capacity
			} else {
				probe.pathBls = p.pathBls
			}
			err := r.sendProbeToRouter(neighbor, probe)
			if err != nil {
				os.Exit(1)
				return fmt.Errorf("router %v handle probe"+
					" %v failed : %v", r.ID, p, err)
			}
		}
		// 发送过就标记为false
		r.BestHopsTable[p.dest].updated = false

		// 找到关于这个dest的路由表信息，我们根据收到的probe和路由表决定要不要更新路由表
	} else {

		// 如果还是上一跳发来的probe，我们无条件更新
		if bestHopEntry.upperHop == p.upper {
			bestHopEntry.dis = p.pathDis + 1
			capacity := r.getLink(bestHopEntry.upperHop).capacity
			if capacity < p.pathBls {
				bestHopEntry.capacity = capacity
			} else {
				bestHopEntry.capacity = p.pathBls
			}
			bestHopEntry.updated = true

			// 不是上一跳发来的probe，那么再分两种情况分析
		} else {

			// 如果跳数更少，则更新
			if bestHopEntry.dis > p.pathDis+1 {
				bestHopEntry.dis = p.pathDis + 1
				bestHopEntry.upperHop = p.upper
				bestHopEntry.capacity = p.pathBls
				bestHopEntry.updated = true

				// 跳数相同，但是capacity增大也更新
			} else if bestHopEntry.dis == p.pathDis+1 &&
				bestHopEntry.capacity < p.pathBls {
				bestHopEntry.upperHop = p.upper
				bestHopEntry.capacity = p.pathBls
				bestHopEntry.updated = true
			}
		}
		lastUpdate, ok := r.ProbeUpdateTable[p.dest]
		if !ok {
			os.Exit(1)
			return fmt.Errorf("cann't find router ID :%v in"+
				" updateTable", p.dest)
		}

		nowTime := time.Now().Unix()
		if nowTime-lastUpdate >= UpdateWindow &&
			bestHopEntry.updated == true {
			for _, neighbor := range r.Neighbors {
				probe := &Probe{
					dest:    p.dest,
					upper:   r.ID,
					pathDis: bestHopEntry.dis,
					pathBls: min(bestHopEntry.capacity,
						r.getLink(neighbor).capacity),
				}
				err := r.sendProbeToRouter(neighbor, probe)
				if err != nil {
					os.Exit(1)
					return fmt.Errorf("router %v handle probe"+
						" %v failed : %v", r.ID, p, err)
				}
			}
			r.ProbeUpdateTable[p.dest] = time.Now().Unix()
			bestHopEntry.updated = false
		}
	}

	return nil
}

func (r *HulaRouter) printBestHopTable() {
	fmt.Printf("router ID: %v table is :%v \n", r.ID, r.BestHopsTable)
	for dest, entry := range r.BestHopsTable {
		fmt.Printf("dest: %v entry is %v\n", dest, entry)
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
		ID:               id,
		ProbeUpdateTable: make(map[RouterID]int64),
		BestHopsTable:    make(map[RouterID]*HopTableEntry),
		Neighbors:        make([]RouterID, 0),
		MessagePool:      make(chan *Probe, BufferSize),
		timer:            time.NewTicker(SendCycle * time.Second),
		wg:               sync.WaitGroup{},
		quit:             make(chan struct{}),
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

// addLink adds a link between two nodes
func addLink(r1, r2 RouterID, capacity int64, nodeBase map[RouterID]*HulaRouter,
	linkBase map[string]*HulaLink) error {
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

	nodeBase[r1].Neighbors = append(nodeBase[r1].Neighbors, r2)
	nodeBase[r2].Neighbors = append(nodeBase[r2].Neighbors, r1)
	return nil
}
