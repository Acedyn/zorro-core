package scheduling

import (
	"testing"
	"time"

	"github.com/Acedyn/zorro-core/internal/client"
	"github.com/Acedyn/zorro-core/internal/context"
	"github.com/life4/genesis/maps"
)

var contextTest = context.Context{
	Plugins: []*context.Plugin{
		{
			Clients: []*client.Client{
				{
					Name:                "bash",
					StartClientTemplate: "{{.Name}}",
				},
				{
					Name:                "cmd",
					StartClientTemplate: "{{.Name}}",
				},
			},
		},
	},
}

// Test client query over client pool
var clientQueryTests = []*client.ClientQuery{
	{
		Name: &[]string{"bash"}[0],
	},
	// {
	// 	Name:    &[]string{"foo"}[0],
	// 	Version: &[]string{"2.3"}[0],
	// 	Pid:     &[]int32{69}[0],
	// },
}

var runningClientPool = map[string]*RegisteredClient{
	"": {
		Client: &client.Client{
			Name:    "foo",
			Version: "2.3",
			Pid:     69,
		},
	},
}

// Fake that a client is being registered
func mockedScheduler() {
	for {
		queuedClients := maps.Values(client.ClientQueue())
		if len(queuedClients) > 0 {
			registerClient(queuedClients[0].Client)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestClientFromQuery(t *testing.T) {
	go mockedScheduler()
	for clientId, runningClient := range runningClientPool {
		ClientPool()[clientId] = runningClient
	}

	for _, clientQueryTest := range clientQueryTests {
		_, err := ClientFromQuery(&contextTest, clientQueryTest)
		if err != nil {
			t.Errorf("An error occured while getting client from query %s: %s", clientQueryTest, err.Error())
			return
		}
	}
}
