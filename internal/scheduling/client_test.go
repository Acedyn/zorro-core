package scheduling

import (
	"testing"

	"github.com/Acedyn/zorro-core/internal/context"
)

var runClientTestWindows = context.Client{
  Name: "cmd",
  RunClientTemplate: "{{.Name}}",
}

var runClientTestLinux = context.Client{
  Name: "bash",
  RunClientTemplate: "{{.Name}}",
}

var clientTestContext = context.Context{
  Plugins: []*context.Plugin{
    {
      Name: "bash",
      Clients: []*context.Client{&runClientTestLinux},
    },
    {
      Name: "cmd",
      Clients: []*context.Client{&runClientTestWindows},
    },
  },
}

func TestRunClient(t *testing.T) {
  runClientTest := &runClientTestLinux
  clientHandle, err := RunClient(runClientTest, &clientTestContext, map[string]string{
    
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

var clientQueryTests = []*ClientQuery{
  {
    Name: &[]string{"bash"}[0],
  },
  {
    Name: &[]string{"foo"}[0],
    Version: &[]string{"2.3"}[0],
    Pid: &[]int32{69}[0],
  },
}

var runningClientPool = []*RunningClient{
  {
    Client: &context.Client{
      Name: "foo",
      Version: "2.3",
    },
    Pid: 69,
  },
}

func TestClientFromQuery(t *testing.T) {
  for _, runningClient := range runningClientPool {
    RunningClients()[int(runningClient.Pid)] = &ClientHandle{
      RunningClient: runningClient,
    }
  }

  for _, clientQueryTest := range clientQueryTests {
    _, err := ClientFromQuery(&clientTestContext, clientQueryTest)
    if err != nil {
      t.Errorf("An error occured while getting client from query %s: %s", clientQueryTest, err.Error())
      return
    }
  }
}
