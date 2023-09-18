package scheduling

import (
	"testing"

	"github.com/Acedyn/zorro-core/internal/tools"
)

type MockedContext struct {}

func (context *MockedContext) AvailableClients() []*Client {
  return []*Client{&runClientTestLinux, &runClientTestWindows}
}

func (context *MockedContext) Environ(b bool) []string {
  return []string{}
}

// Test running a new client's process
var runClientTestWindows = Client{
  Name: "cmd",
  StartClientTemplate: "{{.Name}}",
}

var runClientTestLinux = Client{
  Name: "bash",
  StartClientTemplate: "{{.Name}}",
}

func TestStartClient(t *testing.T) {
  runClientTest := &runClientTestLinux
  clientHandle, err := runClientTest.Start(&MockedContext{}, map[string]string{
    
  })
  if err != nil {
    t.Errorf("An error occured while running client %s: %s", runClientTest, err.Error())
    return
  }
  if err := clientHandle.Process.Kill(); err != nil {
    t.Errorf("An error occured while killing process %d: %s", clientHandle.Process.Pid, err.Error())
    return
  }
}

// Test client query over client pool
var clientQueryTests = []*tools.ClientQuery{
  {
    Name: &[]string{"bash"}[0],
  },
  {
    Name: &[]string{"foo"}[0],
    Version: &[]string{"2.3"}[0],
    Pid: &[]int32{69}[0],
  },
}

var runningClientPool = []*ClientHandle{
  {
    Client: &Client{
      Name: "foo",
      Version: "2.3",
      Pid: 69,
    },
  },
}

func TestClientFromQuery(t *testing.T) {
  for _, runningClient := range runningClientPool {
    ClientPool()[int(runningClient.Client.Pid)] = runningClient
  }

  for _, clientQueryTest := range clientQueryTests {
    _, err := ClientFromQuery(&MockedContext{}, clientQueryTest)
    if err != nil {
      t.Errorf("An error occured while getting client from query %s: %s", clientQueryTest, err.Error())
      return
    }
  }
}
