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

func NewServer() *Server {
	return &Server{
		handlers: make(map[string]Handler),
	}
}

func (server *Server) Connect(url string) error {
	connection, err := amqp.Dial(url)
	if err != nil {
		return err
	}

	server.Connection = connection
	return nil
}

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

	request := &Request{}
	err := json.Unmarshal(delivery.Body, &request)
	if err != nil {
		log.Printf("Error unmarshalling request payload: %v", err)
		return
	}

	var response *Response
	result, err := handler.Handle(request)
	if err != nil {
		response = errorToResponse(err)
	} else {
		response = resultToResponse(result)
	}
	response.Id = request.Id

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

func errorToResponse(err error) *Response {
	return &Response{
		JsonRpc: "2.0",
		Result:  nil,
		Error: &ErrorResponse{
			Code:    500,
			Message: err.Error(),
			Data:    nil,
		},
	}
}

func (server *Server) Drain(queue string, handler Handler) {
	server.handlers[queue] = handler
}

func (server *Server) DrainFunc(queue string, handler func(*Request) (interface{}, error)) {
	server.Drain(queue, HandlerFunc(handler))
}

// Allows ordinary functions to fulfill the Handler interface. Used by the
// DrainFunc method.
type HandlerFunc func(*Request) (interface{}, error)

func (fn HandlerFunc) Handle(req *Request) (interface{}, error) {
	return fn(req)
}
