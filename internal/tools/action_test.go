package tools_test

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Acedyn/zorro-core/internal/context"
	"github.com/Acedyn/zorro-core/internal/network"
	"github.com/Acedyn/zorro-core/internal/tools"
	"github.com/Acedyn/zorro-core/pkg/scheduling"
	_ "github.com/Acedyn/zorro-core/pkg/scheduling/subprocess"

	config_proto "github.com/Acedyn/zorro-proto/zorroprotos/config"
	tools_proto "github.com/Acedyn/zorro-proto/zorroprotos/tools"
	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// Test the resolution of the children that are ready to be traversed
type GetReadyChildrenTest struct {
	Action        *tools_proto.Action
	Pending       map[string]bool
	Completed     []string
	ExpectedReady []string
}

var getReadyChildrenTests = []GetReadyChildrenTest{
	{
		Action: &tools_proto.Action{
			Children: map[string]*tools_proto.ActionChild{
				"1": {
					Child: &tools_proto.ActionChild_Command{
						Command: &tools_proto.Command{},
					},
				},
				"1-1": {
					Child: &tools_proto.ActionChild_Command{
						Command: &tools_proto.Command{},
					},
					Upstream: []string{"1"},
				},
				"2": {
					Child: &tools_proto.ActionChild_Command{
						Command: &tools_proto.Command{},
					},
				},
				"2-1": {
					Child: &tools_proto.ActionChild_Command{
						Command: &tools_proto.Command{},
					},
					Upstream: []string{"2"},
				},
			},
		},
		Pending: map[string]bool{
			"1":   false,
			"1-1": true,
			"2":   false,
			"2-1": true,
		},
		Completed:     []string{"1"},
		ExpectedReady: []string{"1-1"},
	},
	{
		Action: &tools_proto.Action{
			Children: map[string]*tools_proto.ActionChild{
				"1": {
					Child: &tools_proto.ActionChild_Command{
						Command: &tools_proto.Command{},
					},
				},
				"2": {
					Child: &tools_proto.ActionChild_Command{
						Command: &tools_proto.Command{},
					},
				},
				"3": {
					Child: &tools_proto.ActionChild_Command{
						Command: &tools_proto.Command{},
					},
					Upstream: []string{"1", "2"},
				},
			},
		},
		Pending: map[string]bool{
			"1": false,
			"2": false,
			"3": true,
		},
		Completed:     []string{"1"},
		ExpectedReady: []string{},
	},
}

func TestGetReadyChildren(t *testing.T) {
	for _, getReadyChildrenTest := range getReadyChildrenTests {
		action := tools.Action{Action: getReadyChildrenTest.Action}
		readyChildren := action.GetReadyChildren(
			getReadyChildrenTest.Pending,
			getReadyChildrenTest.Completed,
		)
		if !slices.Equal(maps.Keys(readyChildren), getReadyChildrenTest.ExpectedReady) {
			t.Errorf("Incorrect ready children set resolved (received %s, expected %s)",
				maps.Keys(readyChildren),
				getReadyChildrenTest.ExpectedReady,
			)
		}
	}
}

// Test the action traversal order
var actionTraversalTest = tools.Action{
	Action: &tools_proto.Action{
		Base: &tools_proto.ToolBase{
			Name: &[]string{"0"}[0],
		},
		Children: map[string]*tools_proto.ActionChild{
			"00-A": {
				Child: &tools_proto.ActionChild_Action{
					Action: &tools_proto.Action{
						Base: &tools_proto.ToolBase{
							Name: &[]string{"00-A"}[0],
						},
						Children: map[string]*tools_proto.ActionChild{
							"000-A": {
								Child: &tools_proto.ActionChild_Command{
									Command: &tools_proto.Command{
										Base: &tools_proto.ToolBase{
											Name: &[]string{"000-A"}[0],
										},
									},
								},
							},
							"001-A": {
								Child: &tools_proto.ActionChild_Command{
									Command: &tools_proto.Command{
										Base: &tools_proto.ToolBase{
											Name: &[]string{"001-A"}[0],
										},
									},
								},
								Upstream: []string{"000-A"},
							},
							"002-A": {
								Child: &tools_proto.ActionChild_Command{
									Command: &tools_proto.Command{
										Base: &tools_proto.ToolBase{
											Name: &[]string{"002-A"}[0],
										},
									},
								},
								Upstream: []string{"001-A"},
							},
							"002-B": {
								Child: &tools_proto.ActionChild_Action{
									Action: &tools_proto.Action{
										Base: &tools_proto.ToolBase{
											Name: &[]string{"002-B"}[0],
										},
										Children: map[string]*tools_proto.ActionChild{
											"0020-A": {
												Child: &tools_proto.ActionChild_Command{
													Command: &tools_proto.Command{
														Base: &tools_proto.ToolBase{
															Name: &[]string{"0020-A"}[0],
														},
													},
												},
											},
											"0020-B": {
												Child: &tools_proto.ActionChild_Command{
													Command: &tools_proto.Command{
														Base: &tools_proto.ToolBase{
															Name: &[]string{"0020-B"}[0],
														},
													},
												},
											},
										},
									},
								},
								Upstream: []string{"001-A"},
							},
							"002-C": {
								Child: &tools_proto.ActionChild_Command{
									Command: &tools_proto.Command{
										Base: &tools_proto.ToolBase{
											Name: &[]string{"002-C"}[0],
										},
									},
								},
								Upstream: []string{"001-A"},
							},
						},
					},
				},
			},
			"01-A": {
				Child: &tools_proto.ActionChild_Action{
					Action: &tools_proto.Action{
						Base: &tools_proto.ToolBase{
							Name: &[]string{"01-A"}[0],
						},
						Children: map[string]*tools_proto.ActionChild{
							"010-A": {
								Child: &tools_proto.ActionChild_Command{
									Command: &tools_proto.Command{
										Base: &tools_proto.ToolBase{
											Name: &[]string{"010-A"}[0],
										},
									},
								},
							},
							"011-A": {
								Child: &tools_proto.ActionChild_Command{
									Command: &tools_proto.Command{
										Base: &tools_proto.ToolBase{
											Name: &[]string{"011-A"}[0],
										},
									},
								},
								Upstream: []string{"010-A"},
							},
						},
					},
				},
				Upstream: []string{"00-A"},
			},
			"01-B": {
				Child: &tools_proto.ActionChild_Command{
					Command: &tools_proto.Command{
						Base: &tools_proto.ToolBase{
							Name: &[]string{"01-B"}[0],
						},
					},
				},
				Upstream: []string{"00-A"},
			},
			"02-A": {
				Child: &tools_proto.ActionChild_Command{
					Command: &tools_proto.Command{
						Base: &tools_proto.ToolBase{
							Name: &[]string{"02-A"}[0],
						},
					},
				},
				Upstream: []string{"01-A", "01-B"},
			},
		},
	},
}

func TestActionTraversal(t *testing.T) {
	traversalHistory := []string{}
	traversalHistoryMutex := &sync.Mutex{}
	actionTraversalTest.Traverse(func(tool tools.Tool) error {
		traversalHistoryMutex.Lock()
		traversalHistory = append(traversalHistory, tool.GetBase().GetName())
		traversalHistoryMutex.Unlock()
		return nil
	})

	if len(traversalHistory) < 14 {
		t.Errorf(
			"Action traversal did not traversed the full tree (%d children traversed)",
			len(traversalHistory),
		)
	}
	for index, traversedKey := range traversalHistory {
		if index == 0 {
			continue
		}
		traversedKey = strings.Split(traversedKey, "-")[0]
		previousTraversedKey := strings.Split(traversalHistory[index-1], "-")[0]
		minLenght, _ := slices.Min([]int{len(previousTraversedKey), len(traversedKey)})

		if strings.Compare(previousTraversedKey[:minLenght], traversedKey[:minLenght]) == 1 {
			t.Errorf(
				"Action traversal ran in incorrect order (%s < %s)",
				previousTraversedKey,
				traversedKey,
			)
		}
	}
}

func TestActionUnmarshall(t *testing.T) {
	cwdPath, err := os.Getwd()
	if err != nil {
		t.Errorf("Could not get the current working directory: %v", err)
		return
	}
	cwdPath = strings.ReplaceAll(filepath.Dir(filepath.Dir(filepath.Join(cwdPath))), string(filepath.Separator), "/")
	actionPath := strings.ReplaceAll(filepath.Join(cwdPath, "testdata", "actions", "foo.json"), string(filepath.Separator), "/")

	action, err := tools.LoadAction(actionPath)
	if err != nil {
		t.Errorf("An error occured when loading the action at path %s: %v", actionPath, err)
	}

	if action.GetBase().GetName() != "foo" {
		t.Errorf("Expected the name to be foo in the unmarshaled action")
	}

	if action.GetBase().GetLabel() != "Foo" {
		t.Errorf("Expected the label to be Foo in the unmarshaled action")
	}

	if _, inputMessageAExists := action.GetBase().GetInput().GetFields()["input_message_a"]; !inputMessageAExists {
		t.Errorf("Expected an input field at key 'input_message_a' in the unmarshaled action")
	}

	if _, logAExexists := action.GetChildren()["log_a"]; !logAExexists {
		t.Errorf("Expected a child at key 'log_a' in the unmarshaled action")
	}
}

func TestActionExecution(t *testing.T) {
	host := "127.0.0.1"
	port, err := getFreePort()

	if err != nil {
		t.Errorf("Could not get free port: %s", err.Error())
	}

	// Start the server in its own goroutine
	go func() {
		_, grpcStatus := network.GrpcServer()
		if grpcStatus.IsRunning {
			return
		}
		if err := network.ServeGrpc(host, port); err != nil {
			t.Errorf("An error occured while serving GRPC: %s", err.Error())
		}
	}()

	cwdPath, err := os.Getwd()
	if err != nil {
		t.Errorf("Could not get the current working directory: %v", err)
		return
	}
	cwdPath = strings.ReplaceAll(filepath.Dir(filepath.Dir(filepath.Join(cwdPath))), string(filepath.Separator), "/")
	actionPath := strings.ReplaceAll(filepath.Join(cwdPath, "testdata", "actions", "bar.json"), string(filepath.Separator), "/")

	fullPath := strings.ReplaceAll(filepath.Join(cwdPath, "testdata", "plugins"), string(filepath.Separator), "/")
	resolvedContext, err := context.NewContext([]string{"python"}, &config_proto.Config{PluginConfig: &config_proto.PluginConfig{
		Repositories: []*config_proto.RepositoryConfig{
			{
				FileSystemConfig: &config_proto.RepositoryConfig_Os{
					Os: &config_proto.OsFsConfig{
						Directory: fullPath,
					},
				},
			},
		},
	}})
	if err != nil {
		t.Errorf("An error occured when resolving the context: %v", err)
		return
	}

	action, err := tools.LoadAction(actionPath)
	if err != nil {
		t.Errorf("An error occured when loading the action at path %s: %v", actionPath, err)
		return
	}

	action.GetBase().GetInput().GetField("prefixMessage").Update(&tools.Socket{
		&tools_proto.Socket{
			Value: &tools_proto.Socket_Raw{Raw: []byte("\" it's me\"")},
		},
	})
	err = action.Execute(resolvedContext)
	if err != nil {
		t.Errorf("An error occured when executing the action %s: %v", action.GetBase().GetName(), err)
	}

	actionOutput, err := action.GetBase().GetOutput().ResolveRawValue(action)
	if err != nil {
		t.Errorf("Could not resovle the action %s's output: %v", action.GetBase().GetName(), err)
	}

	if string(actionOutput) != "\"DEBUG: hello it's me\"" {
		t.Errorf("Expected action output to be \"hello it's me\", received \"%s\"", string(actionOutput))
	}
}

func init() {
	// Make sure the subprocess scheduler is initialized
	scheduling.InitializeAvailableSchedulers()
	go scheduling.ListenCommandQueries()
}
