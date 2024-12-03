package main

import (
	"flag"
	"websocketReverseProxy/server"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	server.StartServer(*addr)
}
