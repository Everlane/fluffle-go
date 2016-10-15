package fluffle

import (
	"fmt"
)

func ExampleNewClient() {
	client, err := NewClient("amqp://localhost")
	if err != nil {
		panic(err)
	}

	resp, err := client.Call("uppercase", []interface{}{"Hello world."}, "default")
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("uppercase: %s", resp)
	}
}
