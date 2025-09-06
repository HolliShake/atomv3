package runtime

import "fmt"

type AtomType int

const (
	AtomTypeInt AtomType = iota
	AtomTypeNum
	AtomTypeBool
	AtomTypeStr
	AtomTypeNull
	AtomTypeObj
	AtomTypeFunc
	AtomTypeErr
)

type AtomValue struct {
	Type   AtomType
	Value  any
	Next   *AtomValue
	Marked bool
}

func NewAtomValue(atomType AtomType) *AtomValue {
	obj := new(AtomValue)
	obj.Type = atomType
	obj.Marked = false
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

func NewAtomValueNull() *AtomValue {
	obj := NewAtomValue(AtomTypeNull)
	obj.Value = nil
	return obj
}

func NewAtomValueFalse() *AtomValue {
	obj := NewAtomValue(AtomTypeBool)
	obj.Value = false
	return obj
}

func NewAtomValueTrue() *AtomValue {
	obj := NewAtomValue(AtomTypeBool)
	obj.Value = true
	return obj
}

func NewFunction(file, name string, argc int) *AtomValue {
	obj := NewAtomValue(AtomTypeFunc)
	obj.Value = NewAtomCode(file, name, argc)
	return obj
}

func NewAtomValueError(message string) *AtomValue {
	obj := NewAtomValue(AtomTypeErr)
	obj.Value = message
	return obj
}

func (v *AtomValue) String() string {
	if CheckType(v, AtomTypeNull) {
		return "null"
	}
	return fmt.Sprintf("%s: %v", GetTypeString(v), v.Value)
}

func GetTypeString(value *AtomValue) string {
	switch value.Type {
	case AtomTypeInt:
		return "int"
	case AtomTypeNum:
		return "number"
	case AtomTypeBool:
		return "bool"
	case AtomTypeStr:
		return "string"
	case AtomTypeNull:
		return "null"
	case AtomTypeObj:
		return "object"
	case AtomTypeFunc:
		return "function"
	default:
		return "unknown"
	}
}
