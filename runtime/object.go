package runtime

import "unsafe"

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
	return o.hashValueWithVisited(make(map[*AtomObject]bool))
}

func (o *AtomObject) hashValueWithVisited(visited map[*AtomObject]bool) int {
	// Check if we've already visited this object to prevent infinite recursion
	if visited[o] {
		// Return a consistent hash for already-visited objects
		return int(uintptr(unsafe.Pointer(o)))
	}

	// Mark this object as visited
	visited[o] = true

	hash := 0
	for key, value := range o.Elements {
		// Include the key in the hash calculation to ensure different objects
		// with the same values but different keys have different hashes
		keyHash := 0
		for _, char := range key {
			keyHash = keyHash*31 + int(char)
		}

		valueHash := 0
		if CheckType(value, AtomTypeObj) {
			// Use the recursive helper for object values
			valueHash = value.Value.(*AtomObject).hashValueWithVisited(visited)
		} else {
			// For non-object values, use the regular HashValue method
			valueHash = value.HashValue()
		}

		hash = hash*31 + keyHash + valueHash
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
