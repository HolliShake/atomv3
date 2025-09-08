package main

import "fmt"

type AtomScopeType int

const (
	AtomScopeTypeGlobal AtomScopeType = iota
	AtomScopeTypeFunction
	AtomScopeTypeBlock
	AtomScopeTypeLoop
)

type AtomScope struct {
	Parent  *AtomScope
	Type    AtomScopeType
	Symbols map[string]*AtomSymbol
}

func NewAtomScope(parent *AtomScope, scopeType AtomScopeType) *AtomScope {
	symbols := map[string]*AtomSymbol{}
	return &AtomScope{Parent: parent, Type: scopeType, Symbols: symbols}
}

func (s *AtomScope) Captures() []*AtomSymbol {
	captures := []*AtomSymbol{}
	for _, symbol := range s.Symbols {
		if symbol.Capture {
			captures = append(captures, symbol)
		}
	}
	return captures
}

func (s *AtomScope) HasSymbol(name string) bool {
	current := s
	for current != nil {
		if current.HasLocal(name) {
			return true
		}
		current = current.Parent
	}
	return false
}

func (s *AtomScope) HasLocal(name string) bool {
	return s.Symbols[name] != nil && !s.Symbols[name].Capture
}

func (s *AtomScope) HasCapture(name string) bool {
	return s.Symbols[name] != nil && s.Symbols[name].Capture
}

func (s *AtomScope) AddSymbol(symbol *AtomSymbol) {
	if s.HasLocal(symbol.Name) {
		return
	}
	s.Symbols[symbol.Name] = symbol
}

func (s *AtomScope) AddCapture(symbol *AtomSymbol) {
	if s.HasCapture(symbol.Name) {
		return
	}
	s.Symbols[symbol.Name] = symbol
}

func (s *AtomScope) GetSymbol(name string) *AtomSymbol {
	current := s
	for current != nil {
		if current.HasLocal(name) {
			return current.Symbols[name]
		}
		current = current.Parent
	}
	panic(fmt.Sprintf("Symbol %s not found", name))
}

func (s *AtomScope) GetCapture(name string) *AtomSymbol {
	current := s
	for current != nil {
		if current.HasCapture(name) {
			return current.Symbols[name]
		}
		current = current.Parent
	}
	panic(fmt.Sprintf("Capture %s not found", name))
}

func (s *AtomScope) GetCurrentFunction() *AtomScope {
	current := s
	for current != nil {
		if current.Type == AtomScopeTypeFunction {
			return current
		}
		current = current.Parent
	}
	return nil
}

func (s *AtomScope) InSide(scope AtomScopeType, recurse bool) bool {
	current := s
	for current != nil {
		if current.Type == scope {
			return true
		}
		if !recurse {
			break
		}
		current = current.Parent
	}
	return false
}

func (s *AtomScope) Dump() {
	current := s
	level := 0

	fmt.Println("=== Scope Dump ===")

	for current != nil {
		indent := ""
		for i := 0; i < level; i++ {
			indent += "  "
		}

		scopeType := "Unknown"
		switch current.Type {
		case AtomScopeTypeGlobal:
			scopeType = "Global"
		case AtomScopeTypeFunction:
			scopeType = "Function"
		case AtomScopeTypeBlock:
			scopeType = "Block"
		case AtomScopeTypeLoop:
			scopeType = "Loop"
		}

		fmt.Printf("%s[%s Scope] (%d symbols)\n", indent, scopeType, len(current.Symbols))

		if len(current.Symbols) == 0 {
			fmt.Printf("%s  (no symbols)\n", indent)
		} else {
			for _, symbol := range current.Symbols {
				flags := ""
				if symbol.Global {
					flags += "G"
				}
				if symbol.Const {
					flags += "C"
				}
				if symbol.Capture {
					flags += "P"
				}
				if flags == "" {
					flags = "-"
				}

				fmt.Printf("%s  %-12s offset:%-3d flags:[%s]\n",
					indent, symbol.Name, symbol.Offset, flags)
			}
		}

		current = current.Parent
		level++
	}

	fmt.Println("==================")
}
