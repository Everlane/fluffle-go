package fluffle

import (
	"fmt"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
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
		fmt.Printf("uppercase: %s\n", resp)
	}
}

func TestClient(t *testing.T) {
	client, err := NewClient("amqp://localhost")
	require.Nil(t, err)

	resp, err := client.Call("foo", []interface{}{}, "default")
	require.Nil(t, err)
	require.Equal(t, resp, "bar")
}

func TestDecodeResponseWithResult(t *testing.T) {
	id := "abc123"
	expectedResult := "Hello world."
	response := &Response{
		JsonRpc: "2.0",
		Id:      id,
		Result:  expectedResult,
		Error:   nil,
	}

	result, err := NewBareClient(uuid.Nil).DecodeResponse(response)
	require.Nil(t, err)
	require.Equal(t, result, expectedResult)
}

func TestDecodeResponseWithError(t *testing.T) {
	id := "abc123"
	errorResponse := &ErrorResponse{
		Code:    456,
		Message: "Something bad happened.",
	}
	response := &Response{
		JsonRpc: "2.0",
		Id:      id,
		Result:  nil,
		Error:   errorResponse,
	}

	result, err := NewBareClient(uuid.Nil).DecodeResponse(response)
	require.Nil(t, result)
	require.Equal(t, err, errorResponse)
}

func TestDecodeResponseWithNeither(t *testing.T) {
	id := "abc123"
	response := &Response{
		JsonRpc: "2.0",
		Id:      id,
		Result:  nil,
		Error:   nil,
	}

	result, err := NewBareClient(uuid.Nil).DecodeResponse(response)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Missing")
}
