package blobert

import (
	"log"
	"net"

	"golang.org/x/sys/unix"
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

func GetDiskSpace(dir string) uint64 {
	var stat unix.Statfs_t

	err := unix.Statfs(dir, &stat)
	if err != nil {
		panic(err)
	}

	// Available blocks * size per block = available space in bytes
	return stat.Bavail * uint64(stat.Bsize)
}
