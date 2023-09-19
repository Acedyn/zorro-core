package scheduling

import (
  "testing"

	"github.com/Acedyn/zorro-core/internal/client"
	"github.com/Acedyn/zorro-core/internal/context"
)

var contextTest = context.Context{
  Plugins: []*context.Plugin{
    {
      Clients: []*client.Client{
        {
          Name: "bash",
          StartClientTemplate: "{{.Name}}",
        },
        {
          Name: "cmd",
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
  {
    Name: &[]string{"foo"}[0],
    Version: &[]string{"2.3"}[0],
    Pid: &[]int32{69}[0],
  },
}

var runningClientPool = map[string]*RegisteredClient{
  "": {
    Client: &client.Client{
      Name: "foo",
      Version: "2.3",
      Pid: 69,
    },
  },
}

func TestClientFromQuery(t *testing.T) {
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
