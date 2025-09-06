package runtime

type AtomArray struct {
	Elements []*AtomValue
}

func NewAtomArray(elements []*AtomValue) *AtomArray {
	return &AtomArray{Elements: elements}
}

func (a *AtomArray) Get(index int) *AtomValue {
	return a.Elements[index]
}

func (a *AtomArray) Set(index int, value *AtomValue) {
	a.Elements[index] = value
}

func (a *AtomArray) ValidIndex(index int) bool {
	return index >= 0 && index < len(a.Elements)
}

func (a *AtomArray) Len() int {
	return len(a.Elements)
}
