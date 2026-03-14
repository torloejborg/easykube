package cmd

import (
	"github.com/google/uuid"
	"github.com/torloejborg/easykube/pkg/ez"
)

type IOrderedTask interface {
	GetName() string
	GetDependencies() []string
}

type Task struct {
	Name          string
	Description   string
	Dependencies  []string
	SkipCondition func() bool
	Execute       func() error
	graph         *ez.Graph[Task]
}

func (t Task) GetName() string {
	return t.Name
}

func (t Task) GetDependencies() []string {
	return t.Dependencies
}

func (x Task) DependsOn(t Task) error {
	x.graph.AddEdge(t, x)
	return nil
}

func NewTask(graph *ez.Graph[Task], description string, execute func() error) Task {

	u, _ := uuid.NewUUID()

	return Task{
		graph:         graph,
		Name:          u.String(),
		Description:   description,
		Dependencies:  make([]string, 0),
		Execute:       execute,
		SkipCondition: func() bool { return false },
	}
}

func NewTaskWithSkip(graph *ez.Graph[Task], description string, execute func() error, skip func() bool) Task {
	u, _ := uuid.NewUUID()

	return Task{
		graph:         graph,
		Name:          u.String(),
		Description:   description,
		Dependencies:  make([]string, 0),
		Execute:       execute,
		SkipCondition: skip,
	}
}
