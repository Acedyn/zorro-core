package client

import (
	"testing"
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

func TestStartClient(t *testing.T) {
	runClientTest := &runClientTestLinux
	clientHandle, err := runClientTest.Start(&MockedContext{}, map[string]string{})
	if err != nil {
		t.Errorf("An error occured while running client %s: %s", runClientTest, err.Error())
		return
	}
	if err := clientHandle.Process.Kill(); err != nil {
		t.Errorf("An error occured while killing process %d: %s", clientHandle.Process.Pid, err.Error())
		return
	}
}
