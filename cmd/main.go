package main

import (
	"flag"
	"fmt"
	"log"
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

	bb, err := blobert.NewBlobert(blobert.Options{
		Bootstrap: bootstrap,
		Name:      name,
		BaseDir:   "/tmp",
		Peers:     strings.Split(peers, ","),
	})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(bb)

	http.ListenAndServe(":7000", nil)
}
