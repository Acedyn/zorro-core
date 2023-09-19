package client

import (
	"runtime"
	"testing"
	"time"

	"github.com/life4/genesis/maps"
)

type MockedContext struct{}

func (context *MockedContext) AvailableClients() []*Client {
	return []*Client{&runClientTestLinux, &runClientTestWindows}
}

func (context *MockedContext) Environ(b bool) []string {
	return []string{}
}

// Test running a new client's process
var runClientTestWindows = Client{
	Name:                "cmd",
	StartClientTemplate: "{{.Name}}",
}

var runClientTestLinux = Client{
	Name:                "bash",
	StartClientTemplate: "{{.Name}}",
}

// Fake that a client is being registered
func mockedScheduler() {
	for {
		queuedClients := maps.Values(ClientQueue())
		if len(queuedClients) > 0 {
			queuedClients[0].Registration <- nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestStartClient(t *testing.T) {
	go mockedScheduler()

	var runClientTest *Client = nil
	switch runtime.GOOS {
	case "windows":
		runClientTest = &runClientTestWindows
	case "linux":
		runClientTest = &runClientTestLinux
	default:
		runClientTest = &runClientTestLinux
	}

	_, err := runClientTest.Start(&MockedContext{}, map[string]string{})
	if err != nil {
		t.Errorf("An error occured while running client %s: %s", runClientTest, err.Error())
		return
	}
}
