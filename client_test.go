package fluffle

import (
	"fmt"
	"testing"

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
