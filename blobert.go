package blobert

import (
	"log"
	"net"
)

type Blobert struct {
	cluster *Cluster
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("tcp", "golang.org:http")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.TCPAddr)

	return localAddr.IP
}
