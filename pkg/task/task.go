package task

import "gonum.org/v1/gonum/graph"

type TaskHolder struct {
}

type Task struct {
	node      graph.Node
	Name      string
	Operation func() error
	DependsOn []Task
}

func (t *Task) AddDependsOn(dependsOn ...Task) {
	for i := range dependsOn {

		t.DependsOn = append(t.DependsOn, dependsOn[i])
	}
}

func (t *Task) Execute() {

}
