package tools

import (
	"strings"
	"testing"

	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"
)

// Test the resolution of the children that are ready to be traversed
type GetReadyChildrenTest struct {
	Action        *Action
	Pending       map[string]bool
	Completed     []string
	ExpectedReady []string
}

var getReadyChildrenTests = []GetReadyChildrenTest{
	{
		Action: &Action{
			Children: map[string]*ActionChild{
				"1": {
					Child: &ActionChild_Command{
						Command: &Command{},
					},
				},
				"1-1": {
					Child: &ActionChild_Command{
						Command: &Command{},
					},
					Upstream: []string{"1"},
				},
				"2": {
					Child: &ActionChild_Command{
						Command: &Command{},
					},
				},
				"2-1": {
					Child: &ActionChild_Command{
						Command: &Command{},
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
		Action: &Action{
			Children: map[string]*ActionChild{
				"1": {
					Child: &ActionChild_Command{
						Command: &Command{},
					},
				},
				"2": {
					Child: &ActionChild_Command{
						Command: &Command{},
					},
				},
				"3": {
					Child: &ActionChild_Command{
						Command: &Command{},
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
		readyChildren := getReadyChildrenTest.Action.getReadyChildren(
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
var actionTraversalTest = Action{
	Base: &ToolBase{
		Name: &[]string{"0"}[0],
	},
	Children: map[string]*ActionChild{
		"00-A": {
			Child: &ActionChild_Action{
				Action: &Action{
					Base: &ToolBase{
						Name: &[]string{"00-A"}[0],
					},
					Children: map[string]*ActionChild{
						"000-A": {
							Child: &ActionChild_Command{
								Command: &Command{
									Base: &ToolBase{
										Name: &[]string{"000-A"}[0],
									},
								},
							},
						},
						"001-A": {
							Child: &ActionChild_Command{
								Command: &Command{
									Base: &ToolBase{
										Name: &[]string{"001-A"}[0],
									},
								},
							},
							Upstream: []string{"000-A"},
						},
						"002-A": {
							Child: &ActionChild_Command{
								Command: &Command{
									Base: &ToolBase{
										Name: &[]string{"002-A"}[0],
									},
								},
							},
							Upstream: []string{"001-A"},
						},
						"002-B": {
							Child: &ActionChild_Action{
								Action: &Action{
									Base: &ToolBase{
										Name: &[]string{"002-B"}[0],
									},
									Children: map[string]*ActionChild{
										"0020-A": {
											Child: &ActionChild_Command{
												Command: &Command{
													Base: &ToolBase{
														Name: &[]string{"0020-A"}[0],
													},
												},
											},
										},
										"0020-B": {
											Child: &ActionChild_Command{
												Command: &Command{
													Base: &ToolBase{
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
							Child: &ActionChild_Command{
								Command: &Command{
									Base: &ToolBase{
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
			Child: &ActionChild_Action{
				Action: &Action{
					Base: &ToolBase{
						Name: &[]string{"01-A"}[0],
					},
					Children: map[string]*ActionChild{
						"010-A": {
							Child: &ActionChild_Command{
								Command: &Command{
									Base: &ToolBase{
										Name: &[]string{"010-A"}[0],
									},
								},
							},
						},
						"011-A": {
							Child: &ActionChild_Command{
								Command: &Command{
									Base: &ToolBase{
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
			Child: &ActionChild_Command{
				Command: &Command{
					Base: &ToolBase{
						Name: &[]string{"01-B"}[0],
					},
				},
			},
			Upstream: []string{"00-A"},
		},
		"02-A": {
			Child: &ActionChild_Command{
				Command: &Command{
					Base: &ToolBase{
						Name: &[]string{"02-A"}[0],
					},
				},
			},
			Upstream: []string{"01-A", "01-B"},
		},
	},
}

func TestActionTraversal(t *testing.T) {
	traversalHistory := []string{}
	actionTraversalTest.Traverse(func(tool TraversableTool) error {
		traversalHistory = append(traversalHistory, tool.GetBase().GetName())
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
