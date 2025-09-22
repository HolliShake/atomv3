package main

type AtomScopeType int

const (
	AtomScopeTypeGlobal AtomScopeType = iota
	AtomScopeTypeClass
	AtomScopeTypeFunction
	AtomScopeTypeAsyncFunction
	AtomScopeTypeBlock
	AtomScopeTypeLoop
	AtomScopeTypeSingle
)

type AtomScope struct {
	Parent    *AtomScope
	Type      AtomScopeType
	Names     map[string]*AtomSymbol
	Continues []int
	Breaks    []int
}

func NewAtomScope(parent *AtomScope, scopeType AtomScopeType) *AtomScope {
	return &AtomScope{
		Parent:    parent,
		Type:      scopeType,
		Names:     map[string]*AtomSymbol{},
		Continues: []int{},
		Breaks:    []int{},
	}
}

func (s *AtomScope) AddContinue(address int) {
	s.Continues = append(s.Continues, address)
}

func (s *AtomScope) AddBreak(address int) {
	s.Breaks = append(s.Breaks, address)
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

func (s *AtomScope) GetCurrentLoop() *AtomScope {
	current := s
	for current != nil {
		if current.Type == AtomScopeTypeLoop {
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
