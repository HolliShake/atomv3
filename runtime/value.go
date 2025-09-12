package runtime

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"unsafe"
)

type AtomType int

const (
	AtomTypeInt AtomType = iota
	AtomTypeNum
	AtomTypeBool
	AtomTypeStr
	AtomTypeNull
	AtomTypeClass
	AtomTypeClassInstance
	AtomTypeEnum
	AtomTypeObj
	AtomTypeArray
	AtomTypeFunc
	AtomTypeNativeFunc
	AtomTypeErr
)

type AtomValue struct {
	Type   AtomType
	Value  any
	Next   *AtomValue
	Marked bool
}

type NativeFunction func(intereter *AtomInterpreter, argc int)

func NewAtomValue(atomType AtomType) *AtomValue {
	obj := &AtomValue{}
	obj.Type = atomType
	obj.Marked = false
	obj.Next = nil
	obj.Value = nil
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

func NewAtomValueClass(name string, base, proto *AtomValue) *AtomValue {
	obj := NewAtomValue(AtomTypeClass)
	obj.Value = NewAtomClass(name, base, proto)
	return obj
}

func NewAtomValueEnum(elements map[string]*AtomValue) *AtomValue {
	obj := NewAtomValue(AtomTypeEnum)
	obj.Value = NewAtomObject(elements)
	return obj
}

func NewAtomValueObject(elements map[string]*AtomValue) *AtomValue {
	obj := NewAtomValue(AtomTypeObj)
	obj.Value = NewAtomObject(elements)
	return obj
}

func NewAtomValueArray(elements []*AtomValue) *AtomValue {
	obj := NewAtomValue(AtomTypeArray)
	obj.Value = NewAtomArray(elements)
	return obj
}

func NewAtomValueFunction(file, name string, argc int) *AtomValue {
	obj := NewAtomValue(AtomTypeFunc)
	obj.Value = NewAtomCode(file, name, argc)
	return obj
}

func NewAtomValueNativeFunc(nativeFunc NativeFunc) *AtomValue {
	obj := NewAtomValue(AtomTypeNativeFunc)
	obj.Value = nativeFunc
	return obj
}

func NewAtomValueError(message string) *AtomValue {
	obj := NewAtomValue(AtomTypeErr)
	obj.Value = message
	return obj
}

func (v *AtomValue) String() string {
	switch v.Type {
	case AtomTypeNull:
		return "null"

	case AtomTypeFunc:
		return fmt.Sprintf("function %s(...){}", v.Value.(*AtomCode).Name)

	case AtomTypeArray:
		elements := v.Value.(*AtomArray).Elements
		if len(elements) == 0 {
			return "[]"
		}

		// Pre-allocate slice with estimated capacity
		parts := make([]string, 0, len(elements))
		for _, element := range elements {
			if element.Type == AtomTypeStr {
				parts = append(parts, "'"+element.Value.(string)+"'")
			} else {
				parts = append(parts, element.String())
			}
		}
		return "[" + strings.Join(parts, ", ") + "]"

	case AtomTypeObj:
		objElements := v.Value.(*AtomObject).Elements
		if len(objElements) == 0 {
			return "{}"
		}

		// Pre-allocate slice with estimated capacity
		parts := make([]string, 0, len(objElements))
		for keyStr, value := range objElements {
			var valueStr string
			if value.Type == AtomTypeStr {
				valueStr = "'" + value.Value.(string) + "'"
			} else {
				valueStr = value.String()
			}
			parts = append(parts, keyStr+": "+valueStr)
		}
		return "{" + strings.Join(parts, ", ") + "}"

	case AtomTypeNativeFunc:
		nativeFunc := v.Value.(NativeFunc)
		if nativeFunc.Paramc == Variadict {
			return fmt.Sprintf("%s(...){}", nativeFunc.Name)
		}
		if nativeFunc.Paramc == 0 {
			return fmt.Sprintf("%s(){}", nativeFunc.Name)
		}

		// Pre-allocate slice for parameters
		params := make([]string, nativeFunc.Paramc)
		for i := 0; i < nativeFunc.Paramc; i++ {
			params[i] = fmt.Sprintf("$%d", i)
		}
		return fmt.Sprintf("%s(%s){}", nativeFunc.Name, strings.Join(params, ", "))

	case AtomTypeClass:
		atomClass := v.Value.(*AtomClass)
		return fmt.Sprintf("<class.%s />", atomClass.Name)

	case AtomTypeEnum:
		enumElements := v.Value.(*AtomObject).Elements
		if len(enumElements) == 0 {
			return "enum {}"
		}

		// Pre-allocate slice with estimated capacity
		parts := make([]string, 0, len(enumElements))
		for key, value := range enumElements {
			parts = append(parts, key+"="+value.String())
		}
		return "enum {" + strings.Join(parts, ", ") + "}"

	default:
		return fmt.Sprintf("%v", v.Value)
	}
}

func (v *AtomValue) HashValue() int {
	switch v.Type {
	case AtomTypeInt:
		return int(v.Value.(int32))

	case AtomTypeNum:
		// Use math.Float64bits for consistent hashing of float64
		bits := math.Float64bits(v.Value.(float64))
		return int(bits ^ (bits >> 32))

	case AtomTypeBool:
		// Branchless boolean to int conversion
		return int(*(*uint8)(unsafe.Pointer(&v.Value)))

	case AtomTypeStr:
		// Use unsafe string to bytes conversion to avoid allocation
		str := v.Value.(string)
		if len(str) == 0 {
			return 0
		}
		data := unsafe.Slice(unsafe.StringData(str), len(str))
		hash := 0
		for _, b := range data {
			hash = hash*31 + int(b)
		}
		return hash

	case AtomTypeNull:
		return 0

	case AtomTypeArray:
		elements := v.Value.(*AtomArray).Elements
		if len(elements) == 0 {
			return 0
		}
		hash := 0
		for _, element := range elements {
			hash = hash*31 + element.HashValue()
		}
		return hash

	case AtomTypeObj, AtomTypeEnum:
		return v.Value.(*AtomObject).HashValue()

	case AtomTypeFunc:
		return v.Value.(*AtomCode).HashValue()

	case AtomTypeNativeFunc:
		f := v.Value.(NativeFunc)
		return int(reflect.ValueOf(f.Callable).Pointer())

	case AtomTypeErr:
		// Use unsafe string to bytes conversion to avoid allocation
		str := v.Value.(string)
		if len(str) == 0 {
			return 0
		}
		data := unsafe.Slice(unsafe.StringData(str), len(str))
		hash := 0
		for _, b := range data {
			hash = hash*31 + int(b)
		}
		return hash

	default:
		panic("unknown type")
	}
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
	case AtomTypeEnum:
		return "enum"
	case AtomTypeObj:
		return "object"
	case AtomTypeArray:
		return "array"
	case AtomTypeFunc:
		return "function"
	case AtomTypeNativeFunc:
		return "native_function"
	case AtomTypeErr:
		return "error"
	default:
		return "unknown"
	}
}
