package main

type AtomScopeType int

const (
	AtomScopeTypeGlobal AtomScopeType = iota
	AtomScopeTypeClass
	AtomScopeTypeFunction
	AtomScopeTypeAsyncFunction
	AtomScopeTypeNamespace
	AtomScopeTypeBlock
	AtomScopeTypeBlockNoEnv
	AtomScopeTypeLoop
	AtomScopeTypeSingle
)

type AtomScope struct {
	Parent    *AtomScope
	Type      AtomScopeType
	Alias     string
	Names     map[string]*AtomSymbol
	Continues []int
	Breaks    []int
}

func NewAtomScope(parent *AtomScope, scopeType AtomScopeType) *AtomScope {
	return &AtomScope{
		Parent:    parent,
		Type:      scopeType,
		Alias:     "",
		Names:     map[string]*AtomSymbol{},
		Continues: []int{},
		Breaks:    []int{},
	}
}

func NewAtomAliasScope(parent *AtomScope, scopeType AtomScopeType, alias string) *AtomScope {
	return &AtomScope{
		Parent:    parent,
		Type:      scopeType,
		Alias:     alias,
		Names:     map[string]*AtomSymbol{},
		Continues: []int{},
		Breaks:    []int{},
	}
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
