package blobert

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
)

type Blobert struct {
	raft    *raft.Raft
	cluster *ClusterState
}

type Options struct {
	Bootstrap bool
	Name      string
	BaseDir   string
	Peers     []string
}

func NewBlobert(opts Options) (*Blobert, error) {
	fmt.Printf("Creating node ID %s\n", opts.Name)
	ip := GetOutboundIP()
	fmt.Println("My address is", ip.String())

	rconf := raft.DefaultConfig()

	rconf.LocalID = raft.ServerID(ip.String())

	cluster := &ClusterState{}

	ldb, err := boltdb.NewBoltStore(filepath.Join(opts.BaseDir, "logs.dat"))
	if err != nil {
		return nil, err
	}

	sdb, err := boltdb.NewBoltStore(filepath.Join(opts.BaseDir, "stable.dat"))
	if err != nil {
		return nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(opts.BaseDir, "stable.dat"), err)
	}

	fss, err := raft.NewFileSnapshotStore(opts.BaseDir, 3, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, opts.BaseDir, err)
	}

	transport, err := raft.NewTCPTransport(ip.String(), ip, 10, time.Minute, os.Stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	r, err := raft.NewRaft(rconf, cluster, ldb, sdb, fss, transport)
	if err != nil {
		return nil, fmt.Errorf("creating raft: %w", err)
	}

	b := &Blobert{
		raft:    r,
		cluster: cluster,
	}

	if opts.Bootstrap {
		err = r.BootstrapCluster(raft.Configuration{
			Servers: []raft.Server{
				{
					ID:       rconf.LocalID,
					Address:  transport.LocalAddr(),
					Suffrage: raft.Voter,
				},
			},
		}).Error()
		if err != nil {
			log.Println("Error bootstrapping", err)
		} else {
			log.Println("Bootstrapped the cluster")
		}

		time.Sleep(time.Second * 5)

		if b.raft.State() == raft.Leader {
			fmt.Println("I'm the leader!")
			for _, peer := range opts.Peers {
				if peer == "" {
					continue
				}
				err = b.connectPeer(peer)
				if err != nil {
					log.Fatalln(err)
				}
			}
		}
	}

	return b, nil
}

func (b *Blobert) connectPeer(peer string) error {
	ips, err := net.LookupIP(peer)
	if err != nil {
		return fmt.Errorf("connecting to peer %s: %w", peer, err)
	}
	addr := raft.ServerAddress(fmt.Sprintf("%s:%d", ips[0].String(), 3333))
	fmt.Println("Connecting peer", peer, addr)
	return b.raft.AddVoter(raft.ServerID(peer), addr, 0, time.Minute).Error()
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.Addr {
	conn, err := net.Dial("tcp", "golang.org:http")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.TCPAddr)

	localAddr.Port = 3333

	return localAddr
}
