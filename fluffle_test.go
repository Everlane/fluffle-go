package fluffle

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestEncode(t *testing.T) {
	request := &Request{
		JsonRpc: "2.0",
		Id: "abc",
		Method: "def",
		Params: []interface{}{"ghi"},
	}

	body, err := json.Marshal(request)
	require.Nil(t, err)
	require.Equal(t, string(body), `{"jsonrpc":"2.0","id":"abc","method":"def","params":["ghi"]}`)
}
