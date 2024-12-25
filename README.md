### Golang websocket server with redis `echoing` integration

### How to run
1. Setup redis server - [link](https://redis.io/docs/latest/operate/oss_and_stack/install/install-stack/docker/)
2. Download dependencies - `go mod download `
- `go run ./cmd addr=yourport redisAddr=redisHost` - runs main.go with params.
- `go run ./cmd` - without params (8080, 6379) will be used by default
