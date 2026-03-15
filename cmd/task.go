package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/slok/gospinner"
	"github.com/torloejborg/easykube/pkg/ez"
)

type IOrderedTask interface {
	GetName() string
	GetDependencies() []string
}

type TaskContainer struct {
	graph *ez.Graph[Task]
}

func (t TaskContainer) AddTask(task Task) {
	task.graph = t.graph
	t.graph.AppendNode(task)
}

func (t TaskContainer) GetNodes() []Task {
	return t.graph.Nodes
}

func NewTaskContainer() *TaskContainer {
	return &TaskContainer{
		graph: ez.NewGraph[Task](),
	}
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
	return x.graph.AddEdge(t, x)
}

func NewTask(description string, execute func() error) Task {

	u, _ := uuid.NewUUID()

	return Task{
		Name:          u.String(),
		Description:   description,
		Dependencies:  make([]string, 0),
		Execute:       execute,
		SkipCondition: func() bool { return false },
	}
}

func NewTaskWithSkip(description string, execute func() error, skip func() bool) Task {
	u, _ := uuid.NewUUID()

	return Task{
		Name:          u.String(),
		Description:   description,
		Dependencies:  make([]string, 0),
		Execute:       execute,
		SkipCondition: skip,
	}
}

func ExecuteTasks(taskContainer *TaskContainer) {

	checkErr := func(err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	tasks := taskContainer.GetNodes()

	for i := range tasks {

		s, err := gospinner.NewSpinner(gospinner.Dots)
		checkErr(err)
		checkErr(s.Start(fmt.Sprintf("%s", tasks[i].Description)))

		// Execute task
		if !tasks[i].SkipCondition() {
			time.Sleep(5 * time.Millisecond)
			if err := tasks[i].Execute(); err != nil {
				checkErr(s.Fail())
				fmt.Printf("\r%s %s: %s\n", color.RedString("✗"), tasks[i].Description, err.Error())
				break
			} else {
				checkErr(s.Succeed())
			}
		} else {
			color.Set(color.FgHiBlack)
			_ = s.FinishWithMessage("✔", tasks[i].Description)
			color.Unset()
		}
	}

}
