package runtime

type AtomType int

const (
	AtomTypeInt AtomType = iota
	AtomTypeNum
	AtomTypeBool
	AtomTypeStr
	AtomTypeNull
	AtomTypeObj
	AtomTypeFunc
)

type AtomValue struct {
	Type  AtomType
	Value any
	Next  *AtomValue
}

func NewAtomValue(atomType AtomType) *AtomValue {
	obj := new(AtomValue)
	obj.Type = atomType
	return obj
}

func NewAtomValueInt(value int) *AtomValue {
	obj := NewAtomValue(AtomTypeInt)
	obj.Value = int32(value)
	return obj
}

func NewAtomValueNum(value float64) *AtomValue {
	obj := NewAtomValue(AtomTypeNum)
	obj.Value = value
	return obj
}

func NewAtomValueStr(value string) *AtomValue {
	obj := NewAtomValue(AtomTypeStr)
	obj.Value = value
	return obj
}

func NewFunction(file, name string) *AtomValue {
	obj := NewAtomValue(AtomTypeFunc)
	obj.Value = NewCode(file, name)
	return obj
}
