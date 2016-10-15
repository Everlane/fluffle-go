package fluffle

import (
	"fmt"
)

type Response struct {
	JsonRpc string         `json:"jsonrpc"`
	Id      string         `json:"id"`
	Result  interface{}    `json:"result"`
	Error   *ErrorResponse `json:"error"`
}

// Format of an error object within a response object.
type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

type Request struct {
	JsonRpc string        `json:"jsonrpc"`
	Id      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

const DEFAULT_EXCHANGE = ""

func RequestQueueName(name string) string {
	return fmt.Sprintf("fluffle.requests.%s", name)
}

func ResponseQueueName(name string) string {
	return fmt.Sprintf("fluffle.responses.%s", name)
}
