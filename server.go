package fluffle

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type Handler interface {
	Handle(*Request) (interface{}, error)
}

type Server struct {
	Connection *amqp.Connection

	channel  *amqp.Channel
	handlers map[string]Handler
}

// Initialize a new server with all the necessary internal data structures.
func NewServer() *Server {
	return &Server{
		handlers: make(map[string]Handler),
	}
}

// Connect to an AMQP server. This should be called after initializing a
// server with NewServer and before calling Start.
func (server *Server) Connect(url string) error {
	connection, err := amqp.Dial(url)
	if err != nil {
		return err
	}

	server.Connection = connection
	return nil
}

// Declares queues for the configured handlers on the server and starts
// consuming payloads for those queues. This will block the caller goroutine
// until the server's channel is closed (eg. by calling Stop() on the server).
func (server *Server) Start() error {
	channel, err := server.Connection.Channel()
	if err != nil {
		return err
	}
	server.channel = channel

	for queue, handler := range server.handlers {
		durable := false
		autoDelete := false
		exclusive := false
		noWait := false
		queue, err := channel.QueueDeclare(RequestQueueName(queue), durable, autoDelete, exclusive, noWait, nil)
		if err != nil {
			return err
		}

		err = server.consumeQueue(queue.Name, handler)
		if err != nil {
			return err
		}
	}

	closeChan := make(chan *amqp.Error)
	channel.NotifyClose(closeChan)

	var closeErr *amqp.Error = <-closeChan
	if closeErr != nil {
		return closeErr
	}

	return nil
}

// Closes the server's channel. This will make it stop consuming payloads on
// its handlers' queues.
func (server *Server) Stop() error {
	err := server.channel.Close()
	server.channel = nil
	return err
}

func (server *Server) consumeQueue(queue string, handler Handler) error {
	autoAck := false
	exclusive := false
	noLocal := false
	noWait := false
	deliveries, err := server.channel.Consume(queue, "", autoAck, exclusive, noLocal, noWait, nil)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range deliveries {
			server.handleRequest(&delivery, handler)
		}
	}()

	return nil
}

func (server *Server) handleRequest(delivery *amqp.Delivery, handler Handler) {
	delivery.Ack(false)

	var err error
	var response *Response
	var result interface{}

	request := &Request{}
	err = json.Unmarshal(delivery.Body, &request)
	if err != nil {
		log.Printf("Error unmarshalling request payload: %v", err)
		response = errorToResponse(&ParseError{})
		goto publishResponse
	}

	result, err = handler.Handle(request)
	if err != nil {
		response = errorToResponse(WrapError(err))
	} else {
		response = resultToResponse(result)
	}
	response.Id = request.Id

publishResponse:
	server.publishResponse(response, delivery.ReplyTo)
}

func (server *Server) publishResponse(response *Response, replyTo string) {
	body, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling response payload: %v", err)
		return
	}

	mandatory := false
	immediate := false
	routingKey := replyTo
	publishing := amqp.Publishing{
		CorrelationId: response.Id,
		Body:          body,
	}
	err = server.channel.Publish(DEFAULT_EXCHANGE, routingKey, mandatory, immediate, publishing)
	if err != nil {
		log.Printf("Error publishing response: %v", err)
	}
}

func resultToResponse(result interface{}) *Response {
	return &Response{
		JsonRpc: "2.0",
		Result:  result,
		Error:   nil,
	}
}

func errorToResponse(err Error) *Response {
	return &Response{
		JsonRpc: "2.0",
		Result:  nil,
		Error: &ErrorResponse{
			Code:    err.Code(),
			Message: err.Message(),
			Data:    err.Data(),
		},
	}
}

// Add a handler for a given queue.
func (server *Server) Drain(queue string, handler Handler) {
	server.handlers[queue] = handler
}

// Add a function to handle requests on a given queue.
func (server *Server) DrainFunc(queue string, handler func(*Request) (interface{}, error)) {
	server.Drain(queue, HandlerFunc(handler))
}

// Allows ordinary functions to fulfill the Handler interface. Used by the
// DrainFunc method.
type HandlerFunc func(*Request) (interface{}, error)

func (fn HandlerFunc) Handle(req *Request) (interface{}, error) {
	return fn(req)
}
