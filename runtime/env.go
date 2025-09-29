package runtime

import (
	"fmt"
	"strings"
)

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
	// Collect all environments from current to root
	var envs []*AtomEnv
	for current := e; current != nil; current = current.Parent {
		envs = append(envs, current)
	}

	// Print from parent to children (reverse order)
	for i := len(envs) - 1; i >= 0; i-- {
		current := envs[i]
		depth := len(envs) - 1 - i
		indent := strings.Repeat("  ", depth)
		fmt.Printf("%s┌─────────────────────────────────────┐\n", indent)
		fmt.Printf("%s│            ENV (Level %d)            │\n", indent, depth)
		fmt.Printf("%s├─────────────────────────────────────┤\n", indent)
		if len(current.Locals) == 0 {
			fmt.Printf("%s│            (empty)                  │\n", indent)
		} else {
			render := func(value *AtomValue) string {
				switch value.Type {
				case AtomTypeArray:
					return "[array]"
				case AtomTypeObj:
					return "[object]"
				case AtomTypeFunc:
					return "[function]"
				case AtomTypeNativeFunc:
					return "[native function]"
				case AtomTypeMethod:
					return "[method]"
				case AtomTypeNativeMethod:
					return "[native method]"
				case AtomTypeClass:
					return "[class]"
				case AtomTypeClassInstance:
					return "[class instance]"
				default:
					return value.String()
				}
			}
			vars := []string{}
			for name, value := range current.Locals {
				format := fmt.Sprintf("%s = %s", name, render(value))
				vars = append(vars, format)
			}
			sorted := []string{}
			// Bubble sort
			sorted = make([]string, len(vars))
			copy(sorted, vars)
			for i := 0; i < len(sorted); i++ {
				for j := 0; j < len(sorted)-1-i; j++ {
					if sorted[j] > sorted[j+1] {
						sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
					}
				}
			}
			maxChar := 37
			for _, v := range sorted {
				fmt.Printf("%s│ %s", indent, v)
				// white space padding
				paddingValue := maxChar - len(v) - 1
				if paddingValue < 0 {
					paddingValue = 0
				}
				for range paddingValue {
					fmt.Printf(" ")
				}
				fmt.Printf("│\n")
			}
		}
		fmt.Printf("%s└─────────────────────────────────────┘\n", indent)
		if i > 0 {
			fmt.Printf("%s                   │\n", indent)
			fmt.Printf("%s                   ▼\n", indent)
		}
	}
}
