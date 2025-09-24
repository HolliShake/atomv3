package runtime

import (
	"os"
	"path/filepath"
)

type AtomState struct {
	Path          string
	ModuleLookup  map[string]bool
	FunctionTable *AtomStack
	NullValue     *AtomValue
	FalseValue    *AtomValue
	TrueValue     *AtomValue
}

func NewAtomState() *AtomState {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return &AtomState{
		Path:          filepath.Dir(path),
		ModuleLookup:  map[string]bool{},
		FunctionTable: NewAtomStack(),
		NullValue:     NewAtomValueNull(),
		FalseValue:    NewAtomValueFalse(),
		TrueValue:     NewAtomValueTrue(),
	}
}

func (s *AtomState) SaveModule(name string) (exists bool) {
	if _, exists := s.ModuleLookup[name]; exists {
		return exists
	}
	s.ModuleLookup[name] = true
	return false
}

func (s *AtomState) SaveFunction(obj *AtomValue) int {
	s.FunctionTable.Push(obj)
	return s.FunctionTable.Len() - 1
}
