package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/slok/gospinner"
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

func ExecuteTasks(tasks []Task) {

	for i := range tasks {

		s, _ := gospinner.NewSpinner(gospinner.Dots)
		_ = s.Start(fmt.Sprintf("%s", tasks[i].Description))

		// Execute task
		if !tasks[i].SkipCondition() {
			if err := tasks[i].Execute(); err != nil {
				_ = s.Stop()
				fmt.Printf("\r%s %s: %s\n", color.RedString("✗"), tasks[i].Description, err.Error())
				_ = s.Fail()
				break
			} else {
				_ = s.Succeed()
			}
		} else {
			color.Set(color.FgHiBlack, color.CrossedOut)
			_ = s.FinishWithMessage("✔", tasks[i].Description)
			color.Unset()
		}
		_ = s.Stop()

	}

}
