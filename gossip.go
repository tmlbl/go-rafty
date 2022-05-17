package blobert

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

const rpcPort = 3123

type Cluster struct {
	state NodeMap
}

type NodeInfo struct {
	IP net.IP
}

func NewNodeInfo() (NodeInfo, error) {
	addr := GetOutboundIP()
	return NodeInfo{addr}, nil
}

type NodeMap map[string]NodeInfo

func (c *Cluster) NodeInfo(state *NodeMap, merged *NodeMap) error {
	log.Println("NodeInfo called", state)
	m := c.state
	for k, v := range *state {
		m[k] = v
	}
	*merged = m
	return nil
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

	fmt.Println("Got merged", merged)

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
		info, err := NewNodeInfo()
		if err != nil {
			panic(err)
		}
		c.state[info.IP.String()] = info
		peer := c.randomPeer()
		log.Println("Random peer", peer)
		c.exchangeState(peer)
	}
}

func NewCluster(peers []string) (*Cluster, error) {
	c := Cluster{}
	err := rpc.Register(&c)
	rpc.HandleHTTP()
	if err != nil {
		return nil, err
	}
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		return nil, err
	}

	log.Println("Starting RPC server...")

	// Get our own node info
	info, err := NewNodeInfo()
	if err != nil {
		return nil, err
	}
	c.state = NodeMap{
		info.IP.String(): info,
	}

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

	go c.gossipWorker()

	return &c, nil
}
