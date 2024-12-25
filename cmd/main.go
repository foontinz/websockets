package main

import (
	"flag"
	"websocketReverseProxy/server"
	"websocketReverseProxy/sink"

	"github.com/redis/go-redis/v9"
)

var appAddr = flag.String("addr", "localhost:8080", "http service address")
var redisAddr = flag.String("redisAddr", "localhost:6379", "redis service address")

func main() {
	flag.Parse()
	redis := sink.NewRedisClient(redis.NewClient(&redis.Options{Addr: *redisAddr}))
	server.StartServer(*appAddr, redis)
}
