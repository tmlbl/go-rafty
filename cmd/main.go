package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/tmlbl/blobert"
)

var peers string
var name string
var directory string

func main() {
	flag.StringVar(&peers, "peers", "", "One or more comma-separated cluster peer DNS names")
	flag.StringVar(&name, "name", "", "The DNS name of this instance")
	flag.StringVar(&directory, "dir", "/tmp/blobert", "Directory to store data")
	flag.Parse()

	cluster, err := blobert.NewCluster(directory, strings.Split(peers, ","))
	if err != nil {
		panic(err)
	}

	log.Println(cluster)

	http.ListenAndServe(":7000", nil)
}
