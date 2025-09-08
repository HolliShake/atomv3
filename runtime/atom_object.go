package runtime

type AtomObject struct {
	Freeze   bool
	Elements map[string]*AtomValue
}

func NewAtomObject(elements map[string]*AtomValue) *AtomObject {
	return &AtomObject{Elements: elements, Freeze: false}
}

func (o *AtomObject) Get(key string) *AtomValue {
	return o.Elements[key]
}

func (o *AtomObject) Set(key string, value *AtomValue) {
	o.Elements[key] = value
}

func (o *AtomObject) Len() int {
	return len(o.Elements)
}
