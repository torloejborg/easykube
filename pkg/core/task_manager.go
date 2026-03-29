package core

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/google/uuid"
)

type IOrderedTask interface {
	GetName() string
	GetDependencies() []string
}

type TaskContainer struct {
	graph *Graph[Task]
	mu    sync.Mutex // Mutex to synchronize spinner access
}

func (t *TaskContainer) AddTask(task Task) {
	task.graph = t.graph
	t.graph.AppendNode(task)
}

func (t *TaskContainer) GetNodes() Stack[Task] {
	var s Stack[Task]

	slices.Reverse(t.graph.Nodes)
	for _, node := range t.graph.Nodes {
		s.Push(node)
	}

	return s
}

func NewTaskContainer() *TaskContainer {
	return &TaskContainer{
		graph: NewGraph[Task](),
	}
}

type Task struct {
	Name          string
	Description   string
	Dependencies  []string
	SkipCondition func() bool
	Execute       func() error
	graph         *Graph[Task]
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
	tasks := taskContainer.GetNodes()

	for !tasks.IsEmpty() {
		currentTask, _ := tasks.Pop()

		// Create a new spinner
		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Prefix = currentTask.Description + " "
		s.Start()

		// Execute task
		if !currentTask.SkipCondition() {
			if err := currentTask.Execute(); err != nil {
				s.Stop()
				fmt.Printf("\r%s %s: %s\n", color.RedString("✗"), currentTask.Description, err.Error())
				break
			} else {
				s.Stop()
				fmt.Printf("\r%s %s\n", color.GreenString("✓"), currentTask.Description)
			}
		} else {
			s.Stop()
			// Print skipped task in grey
			fmt.Printf("\r%s\n", color.New(color.FgHiBlack).Sprintf("✔ %s", currentTask.Description))
		}
	}
}
