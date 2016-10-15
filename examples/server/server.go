package main

import (
	"github.com/Everlane/fluffle-go"
)

func main() {
	server := fluffle.NewServer()
	err := server.Connect("amqp://localhost")
	if err != nil {
		panic(err)
	}

	server.DrainFunc("default", func(req *fluffle.Request) (interface{}, error) {
		return "Hello world!", nil
	})

	err = server.Start()
	if err != nil {
		panic(err)
	}
}
