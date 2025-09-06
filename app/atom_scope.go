package main

import "fmt"

type AtomScopeType int

const (
	AtomScopeTypeGlobal AtomScopeType = iota
	AtomScopeTypeLocal
	AtomScopeTypeFunction
	AtomScopeTypeBlock
	AtomScopeTypeLoop
)

type AtomScope struct {
	Parent   *AtomScope
	Type     AtomScopeType
	Symbols  map[string]*AtomSymbol
	Captures map[string]*AtomSymbol
}

func NewAtomScope(parent *AtomScope, scopeType AtomScopeType) *AtomScope {
	symbols := make(map[string]*AtomSymbol)
	captures := make(map[string]*AtomSymbol)
	return &AtomScope{Parent: parent, Type: scopeType, Symbols: symbols, Captures: captures}
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
	return s.Symbols[name] != nil
}

func (s *AtomScope) HasCapture(name string) bool {
	return s.Captures[name] != nil
}

func (s *AtomScope) AddSymbol(symbol *AtomSymbol) {
	s.Symbols[symbol.Name] = symbol
}

func (s *AtomScope) AddCapture(symbol *AtomSymbol) {
	if s.HasCapture(symbol.Name) {
		return
	}
	s.Captures[symbol.Name] = symbol
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
	return s.Captures[name]
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

func (s *AtomScope) InSide(scope AtomScopeType) bool {
	current := s
	for current != nil {
		if current.Type == scope {
			return true
		}
		current = current.Parent
	}
	return false
}
