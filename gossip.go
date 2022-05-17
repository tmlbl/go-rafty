package blobert

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

const rpcPort = 3123

type Cluster struct {
	dir   string
	state NodeMap
}

type NodeInfo struct {
	Time      time.Time
	IP        net.IP
	DiskSpace uint64
}

func (c *Cluster) selfNodeInfo() (NodeInfo, error) {
	return NodeInfo{
		Time:      time.Now(),
		IP:        GetOutboundIP(),
		DiskSpace: GetDiskSpace(c.dir),
	}, nil
}

type NodeMap map[string]NodeInfo

func (c *Cluster) NodeInfo(state *NodeMap, merged *NodeMap) error {
	m := c.state
	for k, v := range *state {
		if _, ok := m[k]; !ok {
			m[k] = v
		} else {
			if m[k].Time.Before(v.Time) {
				log.Println("Received newer info for", v.IP)
				m[k] = v
			}
		}
	}
	*merged = m
	return nil
}

func (c *Cluster) printPeers() {
	for _, peer := range c.state {
		fmt.Println(peer.IP, "disk space:", peer.DiskSpace)
	}
}

func (c *Cluster) exchangeState(peer string) error {
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", peer, rpcPort))
	if err != nil {
		return err
	}
	var merged NodeMap
	err = client.Call("Cluster.NodeInfo", &c.state, &merged)
	if err != nil {
		return err
	}

	c.state = merged
	c.printPeers()
	return nil
}

func (c *Cluster) randomPeer() string {
	i := rand.Intn(len(c.state))
	x := 0
	for peer := range c.state {
		if i == x {
			return peer
		}
		x++
	}
	return ""
}

func (c *Cluster) gossipWorker() {
	for {
		time.Sleep(time.Second * 3)
		info, err := c.selfNodeInfo()
		if err != nil {
			panic(err)
		}
		c.state[info.IP.String()] = info
		peer := c.randomPeer()
		// log.Println("Random peer", peer)
		c.exchangeState(peer)
	}
}

func NewCluster(dir string, peers []string) (*Cluster, error) {
	c := Cluster{
		dir: dir,
	}
	err := rpc.Register(&c)
	rpc.HandleHTTP()
	if err != nil {
		return nil, err
	}
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		return nil, err
	}

	// Ensure data directory exists
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		log.Println("Creatinf data directory", dir)
		err = os.MkdirAll(dir, os.ModeDir)
		if err != nil {
			return nil, err
		}
	}

	// Get our own node info
	info, err := c.selfNodeInfo()
	if err != nil {
		return nil, err
	}
	c.state = NodeMap{
		info.IP.String(): info,
	}

	log.Println("Starting RPC server...")
	go http.Serve(l, nil)

	// Bootstrap
	for _, peer := range peers {
		time.Sleep(time.Second * 3)
		log.Printf("Dialing peer %s...", peer)
		err = c.exchangeState(peer)
		if err != nil {
			return nil, err
		}
	}

	log.Println("Starting gossip worker...")
	go c.gossipWorker()

	return &c, nil
}
