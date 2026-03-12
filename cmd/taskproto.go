package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

type ifacetaskProto interface {
	GetName() string
	GetDependencies() []string
}

type taskProto struct {
	Name         string
	Dependencies []string
	Execute      func() error
}

func (r taskProto) GetName() string {
	return r.Name
}

func (r taskProto) GetDependencies() []string {
	return r.Dependencies
}

var taskProtoCmd = &cobra.Command{
	Use:   "taskproto",
	Short: "demo",
	Long:  `demo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		g := ez.NewGraph[taskProto]()

		t1 := &taskProto{
			Name:         "write configuration",
			Dependencies: make([]string, 0),
			Execute: func() error {
				fmt.Print("Writing configuration... ")
				return nil
			},
		}

		t2 := &taskProto{
			Name:         "provision container",
			Dependencies: make([]string, 0),
			Execute: func() error {
				fmt.Print("Provisioning container... ")
				return nil
			},
		}

		g.AppendNode(*t1)
		g.AppendNode(*t2)
		_ = g.AddEdge(*t1, *t2)

		order, err := g.TopologicalSort()

		if err != nil {
			return err
		} else {
			for _, node := range order {
				e := node.Execute()
				if e != nil {
					return e
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(taskProtoCmd)
}
