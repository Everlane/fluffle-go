package fluffle

import (
	"encoding/json"
	"sync"

	"github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type Client struct {
	UUID          uuid.UUID
	Connection    *amqp.Connection
	Channel       *amqp.Channel
	ResponseQueue *amqp.Queue

	pendingResponses map[string]*pendingResponse
}

type pendingResponse struct {
	cond    *sync.Cond
	payload *Response
}

// Creates a client with the given UUID and initializes its internal data
// structures. Does not setup connections or perform any network operations.
// You will almost always want to use NewClient.
func NewBareClient(uuid uuid.UUID) *Client {
	return &Client{
		UUID:             uuid,
		pendingResponses: make(map[string]*pendingResponse),
	}
}

// Create a client, connect it to the given AMQP server, and setup a queue
// to receive responses on.
func NewClient(url string) (*Client, error) {
	uuid := uuid.NewV1()
	client := NewBareClient(uuid)

	connection, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	client.Connection = connection

	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}
	client.Channel = channel

	err = client.SetupResponseQueue()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Declares the response queue and starts consuming (listening) deliveries
// from it. This is normally the final step in setting up a usable client.
func (c *Client) SetupResponseQueue() error {
	durable := false
	autoDelete := false
	exclusive := true
	noWait := false
	queue, err := c.Channel.QueueDeclare(ResponseQueueName(c.UUID.String()), durable, autoDelete, exclusive, noWait, nil)
	if err != nil {
		return err
	}
	c.ResponseQueue = &queue

	autoAck := false
	noLocal := false
	deliveries, err := c.Channel.Consume(queue.Name, "", autoAck, exclusive, noLocal, noWait, nil)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range deliveries {
			c.handleReply(&delivery)
		}
	}()
	return nil
}

func (c *Client) handleReply(delivery *amqp.Delivery) {
	delivery.Ack(false)

	payload := &Response{}
	json.Unmarshal(delivery.Body, &payload)

	id := payload.Id

	pendingResponse := c.pendingResponses[id]
	pendingResponse.payload = payload
	pendingResponse.cond.Signal()
}

// Call a remote method over JSON-RPC and return its response. This will block
// the goroutine on which it is called.
func (c *Client) Call(method string, params []interface{}, queue string) (interface{}, error) {
	id := uuid.NewV4().String()

	request := &Request{
		JsonRpc: "2.0",
		Id:      id,
		Method:  method,
		Params:  params,
	}

	response, err := c.PublishAndWait(request, queue)
	if err != nil {
		return nil, err
	}

	return c.DecodeResponse(response)
}

// Publishes a request onto the given queue, then waits for a response to that
// message on the client's response queue.
func (c *Client) PublishAndWait(payload *Request, queue string) (*Response, error) {
	pendingResponse := &pendingResponse{
		cond: sync.NewCond(new(sync.Mutex)),
	}
	c.pendingResponses[payload.Id] = pendingResponse
	defer delete(c.pendingResponses, payload.Id)

	err := c.Publish(payload, queue)
	if err != nil {
		return nil, err
	}

	// Block this goroutine until the cond is signaled
	pendingResponse.cond.Wait()

	// Once we're signaled that means the payload is present
	return pendingResponse.payload, nil
}

// payload: JSON-RPC request payload to be sent
//
// queue: Queue on which to send the request
func (c *Client) Publish(payload *Request, queue string) error {
	routingKey := RequestQueueName(queue)
	correlationId := payload.Id
	replyTo := c.ResponseQueue.Name

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	mandatory := false
	immediate := false
	publishing := amqp.Publishing{
		CorrelationId: correlationId,
		ReplyTo:       replyTo,
		Body:          body,
	}
	return c.Channel.Publish(DEFAULT_EXCHANGE, routingKey, mandatory, immediate, publishing)
}

// Figure out what was in the response payload: was it a result, an error,
// or unknown?
func (c *Client) DecodeResponse(payload *Response) (interface{}, error) {
	if payload.Result != nil {
		return payload.Result, nil
	}
	if payload.Error != nil {
		return nil, payload.Error
	}
	return nil, &ErrorResponse{
		Code:    0,
		Message: "Missing both `result' and `error' on Response object",
		Data:    nil,
	}
}
