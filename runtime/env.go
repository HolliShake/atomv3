package runtime

import "fmt"

type AtomEnv struct {
	Parent *AtomEnv
	Locals map[string]*AtomValue
}

func NewAtomEnv(parent *AtomEnv) *AtomEnv {
	return &AtomEnv{
		Parent: parent,
		Locals: map[string]*AtomValue{},
	}
}

func (e *AtomEnv) Has(name string) bool {
	for current := e; current != nil; current = current.Parent {
		if _, ok := current.Locals[name]; ok {
			return true
		}
	}
	return false
}

func (e *AtomEnv) Get(name string) *AtomValue {
	// Propagate to parent
	for current := e; current != nil; current = current.Parent {
		if _, ok := current.Locals[name]; ok {
			return current.Locals[name]
		}
	}
	panic("Not found!!!")
}

func (e *AtomEnv) Put(name string, value *AtomValue) {
	e.Locals[name] = value
}

func (e *AtomEnv) Set(name string, value *AtomValue) {
	// Propagate to parent
	for current := e; current != nil; current = current.Parent {
		if _, ok := current.Locals[name]; ok {
			current.Locals[name] = value
			return
		}
	}
	panic("Not found!!!")
}

func (e *AtomEnv) Dump() {
	fmt.Print("Env: ")
	for name := range e.Locals {
		fmt.Printf("%s, ", name)
	}
	fmt.Println()
}
