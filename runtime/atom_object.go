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

func (o *AtomObject) HashValue() int {
	hash := 0
	for _, value := range o.Elements {
		hash = hash*31 + value.HashValue()
	}
	return hash
}

func (o *AtomObject) ContainsValue(object *AtomValue) bool {
	for _, value := range o.Elements {
		for _, objectValue := range object.Value.(*AtomObject).Elements {
			if value.HashValue() == objectValue.HashValue() {
				return true
			}
		}
	}
	return false
}
