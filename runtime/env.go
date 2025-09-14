package runtime

import "fmt"

type AtomEnv struct {
	Parent *AtomEnv
	Locals map[string]*Variable
}

func NewAtomEnv(parent *AtomEnv) *AtomEnv {
	return &AtomEnv{
		Parent: parent,
		Locals: map[string]*Variable{},
	}
}

func (e *AtomEnv) Has(name string) bool {
	for current := e; current != nil; current = current.Parent {
		if _, exists := current.Locals[name]; exists {
			return true
		}
	}
	return false
}

func (e *AtomEnv) Local(name string) bool {
	return e.Locals[name] != nil
}

func (e *AtomEnv) New(name string, global bool, constant bool, value *AtomValue) error {
	if entry, exists := e.Locals[name]; exists {
		if entry.Const {
			return fmt.Errorf("variable %s is constant", name)
		}
		entry.Value = value
	} else {
		e.Locals[name] = &Variable{
			Name:   name,
			Global: global,
			Const:  constant,
			Value:  value,
		}
	}
	return nil
}

func (e *AtomEnv) Lookup(index string) (*AtomValue, error) {
	for current := e; current != nil; current = current.Parent {
		if entry, exists := current.Locals[index]; exists {
			return entry.Value, nil
		}
	}
	return nil, fmt.Errorf("variable %s not found", index)
}

func (e *AtomEnv) Store(index string, value *AtomValue) (bool, error) {
	for current := e; current != nil; current = current.Parent {
		if entry, exists := current.Locals[index]; exists {
			if entry.Const {
				return false, fmt.Errorf("variable %s is constant", index)
			}
			entry.Value = value
			return true, nil
		}
	}
	return false, fmt.Errorf("variable %s not found", index)
}
