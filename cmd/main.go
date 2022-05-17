package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/tmlbl/blobert"
)

var peers string
var name string
var bootstrap bool

func main() {
	flag.StringVar(&peers, "peers", "", "One or more comma-separated cluster peer DNS names")
	flag.StringVar(&name, "name", "", "The DNS name of this instance")
	flag.BoolVar(&bootstrap, "bootstrap", false, "Whether this is the cluster leader")
	flag.Parse()

	addr := blobert.GetOutboundIP()
	fmt.Println("Hello from", addr)

	cluster, err := blobert.NewCluster(strings.Split(peers, ","))
	if err != nil {
		panic(err)
	}
	fmt.Println(cluster)

	http.ListenAndServe(":7000", nil)
}
