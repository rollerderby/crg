package main

import (
	"flag"

	"github.com/rollerderby/crg/server"
)

var port int

func init() {
	flag.IntVar(&port, "port", 8000, "Server Port")
}

func main() {
	flag.Parse()
	server.Start(version, uint16(port))
}
