# fluffle-go

[![Build Status](https://travis-ci.org/Everlane/fluffle-go.svg?branch=master)](https://travis-ci.org/Everlane/fluffle-go)

This is a Go implementation of the [Fluffle][] design for JSON-RPC over RabbitMQ. It is built for interoperability with the Ruby implementation.

[Fluffle]: https://github.com/Everlane/fluffle

**Note**: Due to the design of the JSON-RPC protocol it is unfeasible to safely implement [batch requests][] on clients in a thread-safe way. Therefore batch requests are not implemented in the client.

[batch requests]: http://www.jsonrpc.org/specification#batch

## Examples

A client is initialized with the URL string to the AMQP server. Calls can then be made to a queue on that server. The `Call` method will block until a response is received or times out (5 seconds).

```go
import "github.com/Everlane/fluffle-go"

client, err := fluffle.NewClient("amqp://localhost")
if err != nil {
	panic(err)
}

resp, err := client.Call("uppercase", []interface{}{"Hello world"}, "default")
if err != nil {
	panic(err)
} else {
	fmt.Printf("%s\n", resp) // -> HELLO WORLD
}
```

Servers are initialized in an "empty" state; they then have their handlers and AMQP connection explicitly set up before calling `Start` to begin blocking for requests.

```go
server := fluffle.NewServer()
err := server.Connect("amqp://localhost")
if err != nil {
	panic(err)
}

server.DrainFunc("default", func(req *fluffle.Request) (interface{}, error) {
	param := req.Params[0].(string)
	return strings.ToUpper(param), nil
})

// This will block the current goroutine
err = server.Start()
if err != nil {
	panic(err)
}
```

## License

Released under the MIT license, see [LICENSE](LICENSE) for details.
