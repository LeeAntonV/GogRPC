package main

import (
	"fmt"
	"gRPC/internal/config"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println(cfg)

	//TODO: Initialize logger

	//TODO: Initialize app

	//TODO: start GRPC-server
}
