package ez

import (
	"fmt"
	"os"
)

type DirStack struct {
	data []string
}

func PushDir(dir string) {
	fmt.Println("pushdir " + dir)
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	dirstack.push(dir)
}

func PopDir() {
	dir, result := dirstack.pop()
	if result {
		//fmt.Println("popped dir " + dir)
		err := os.Chdir(dir)
		if err != nil {
			panic(err)
		}
	}
}

func (s *DirStack) push(v string) {
	s.data = append(s.data, v)
}

func (s *DirStack) pop() (string, bool) {
	if len(s.data) == 0 {
		var zero string
		return zero, false
	}
	val := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return val, true
}

func (s *DirStack) peek() (string, bool) {
	if len(s.data) == 0 {
		var zero string
		return zero, false
	}
	return s.data[len(s.data)-1], true
}

func (s *DirStack) len() int {
	return len(s.data)
}

func (s *DirStack) empty() bool {
	return len(s.data) == 0
}
