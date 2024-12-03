package main

import (
	"flag"
	"log"
	"net/http"
	"websocketReverseProxy/server"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()

	http.HandleFunc("/ws", server.HandleConnections)

	log.Println("Starting websocket server")
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("Errored while listening:", err)
	}
}
