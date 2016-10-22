# fluffle-go

[![Build Status](https://travis-ci.org/Everlane/fluffle-go.svg?branch=master)](https://travis-ci.org/Everlane/fluffle-go)

This is a Go implementation of the [Fluffle][] design for JSON-RPC over RabbitMQ. It is designed for interoperability with the Ruby implementation.

[Fluffle]: https://github.com/Everlane/fluffle

**Note**: Due to the design of the JSON-RPC protocol it is unfeasible to safely implement [batch requests][] on clients in a thread-safe way. Therefore batch requests are not implemented in the client.

[batch requests]: http://www.jsonrpc.org/specification#batch

## License

Released under the MIT license, see [LICENSE](LICENSE) for details.
