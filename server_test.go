package fluffle

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func echoDrainFunc(req *Request) (interface{}, error) {
	params := make([]string, len(req.Params))
	for i, param := range req.Params {
		params[i] = fmt.Sprintf("%#v", param)
	}
	resp := fmt.Sprintf("%v(%v)", req.Method, strings.Join(params, ", "))
	return resp, nil
}

func TestBasicServer(t *testing.T) {
	queue := "test.1"

	server := NewServer()
	err := server.Connect("amqp://localhost")
	require.Nil(t, err)

	server.DrainFunc(queue, echoDrainFunc)

	go func() {
		err := server.Start()
		require.Nil(t, err)
	}()

	client, err := NewClient("amqp://localhost")
	require.Nil(t, err)

	resp, err := client.Call("foo", []interface{}{"bar"}, queue)
	require.Nil(t, err)
	require.Equal(t, resp, `foo("bar")`)

	server.Stop()
}
