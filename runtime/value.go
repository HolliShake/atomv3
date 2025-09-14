package runtime

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
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
	AtomTypeMethod
	AtomTypeFunc
	AtomTypeNativeFunc
	AtomTypeErr
	AtomTypePromise
)

// String builder pool for memory efficiency
var StringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// Pre-allocated common strings to avoid allocations
const (
	NullStr               = "null"
	EmptyObjStr           = "{}"
	EmptyArrStr           = "[]"
	EmptyEnumStr          = "enum {}"
	EmptyClassInstanceStr = " {\n}"
	TrueStr               = "true"
	FalseStr              = "false"
	SelfRefStr            = "[self]"
)

type AtomValue struct {
	Type  AtomType
	Value any
}

type NativeFunction func(intereter *AtomInterpreter, argc int)

func NewAtomValue(atomType AtomType) *AtomValue {
	obj := &AtomValue{}
	obj.Type = atomType
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

func NewAtomValueClassInstance(prototype, property *AtomValue) *AtomValue {
	obj := NewAtomValue(AtomTypeClassInstance)
	obj.Value = NewAtomClassInstance(prototype, property)
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

// Bound method
func NewAtomValueMethod(this *AtomValue, fn *AtomValue) *AtomValue {
	obj := NewAtomValue(AtomTypeMethod)
	obj.Value = NewAtomMethod(this, fn)
	return obj
}

func NewAtomValueFunction(file, name string, async bool, argc int) *AtomValue {
	obj := NewAtomValue(AtomTypeFunc)
	obj.Value = NewAtomCode(file, name, async, argc)
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

func NewAtomValuePromise(state PromiseState, value *AtomValue) *AtomValue {
	obj := NewAtomValue(AtomTypePromise)
	obj.Value = NewAtomPromise(state, value)
	return obj
}

func (v *AtomValue) String() string {
	return v.stringWithVisited(make(map[uintptr]bool))
}

func (v *AtomValue) stringWithVisited(visited map[uintptr]bool) string {
	switch v.Type {
	case AtomTypeInt:
		// Fast path: direct conversion without fmt.Sprintf
		return strconv.FormatInt(int64(v.Value.(int32)), 10)

	case AtomTypeNum:
		// Fast path: direct conversion without fmt.Sprintf
		return strconv.FormatFloat(v.Value.(float64), 'g', -1, 64)

	case AtomTypeBool:
		// Fast path: pre-allocated strings
		if v.Value.(bool) {
			return TrueStr
		}
		return FalseStr

	case AtomTypeStr:
		// Fast path: direct return
		return v.Value.(string)

	case AtomTypeNull:
		// Fast path: pre-allocated string
		return NullStr

	case AtomTypeClass:
		// Optimized: avoid fmt.Sprintf, use helper function
		atomClass := v.Value.(*AtomClass)
		return BuildString(func(b *strings.Builder) {
			b.WriteString("<class.")
			b.WriteString(atomClass.Name)
			b.WriteString(" />")
		})

	case AtomTypeClassInstance:
		// Check for self-reference
		ptr := uintptr(unsafe.Pointer(v))
		if visited[ptr] {
			return SelfRefStr
		}
		visited[ptr] = true
		defer delete(visited, ptr)

		classInstance := v.Value.(*AtomClassInstance)
		prototype := classInstance.Prototype.Value.(*AtomClass)
		properties := classInstance.Property.Value.(*AtomObject).Elements

		if len(properties) == 0 {
			// Fast path for empty class instance
			builder := StringBuilderPool.Get().(*strings.Builder)
			builder.Reset()
			builder.WriteString(prototype.Name)
			builder.WriteString(EmptyClassInstanceStr)
			result := builder.String()
			StringBuilderPool.Put(builder)
			return result
		}

		// Use string builder for complex formatting
		builder := StringBuilderPool.Get().(*strings.Builder)
		builder.Reset()
		builder.WriteString(prototype.Name)
		builder.WriteString(" {\n")

		// Process properties with optimized string building
		first := true
		for key, value := range properties {
			if !first {
				builder.WriteByte('\n')
			}
			first = false

			builder.WriteString("  ")
			builder.WriteString(key)
			builder.WriteString(": ")
			builder.WriteString(valueToStringWithVisited(value, visited))
			builder.WriteByte(',')
		}
		builder.WriteString("\n}")
		result := builder.String()
		StringBuilderPool.Put(builder)
		return result

	case AtomTypeEnum:
		// Check for self-reference
		ptr := uintptr(unsafe.Pointer(v))
		if visited[ptr] {
			return SelfRefStr
		}
		visited[ptr] = true
		defer delete(visited, ptr)

		enumElements := v.Value.(*AtomObject).Elements
		if len(enumElements) == 0 {
			return EmptyEnumStr
		}

		// Use string builder for enum formatting
		builder := StringBuilderPool.Get().(*strings.Builder)
		builder.Reset()
		builder.WriteString("enum {")

		first := true
		for key, value := range enumElements {
			if !first {
				builder.WriteString(", ")
			}
			first = false
			builder.WriteString(key)
			builder.WriteByte('=')
			builder.WriteString(valueToStringWithVisited(value, visited))
		}
		builder.WriteByte('}')
		result := builder.String()
		StringBuilderPool.Put(builder)
		return result

	case AtomTypeObj:
		// Check for self-reference
		ptr := uintptr(unsafe.Pointer(v))
		if visited[ptr] {
			return SelfRefStr
		}
		visited[ptr] = true
		defer delete(visited, ptr)

		objElements := v.Value.(*AtomObject).Elements
		if len(objElements) == 0 {
			return EmptyObjStr
		}

		// Use string builder for object formatting
		builder := StringBuilderPool.Get().(*strings.Builder)
		builder.Reset()
		builder.WriteByte('{')

		first := true
		for keyStr, value := range objElements {
			if !first {
				builder.WriteString(", ")
			}
			first = false
			builder.WriteString(keyStr)
			builder.WriteString(": ")
			builder.WriteString(valueToStringWithVisited(value, visited))
		}
		builder.WriteByte('}')
		result := builder.String()
		StringBuilderPool.Put(builder)
		return result

	case AtomTypeArray:
		// Check for self-reference
		ptr := uintptr(unsafe.Pointer(v))
		if visited[ptr] {
			return SelfRefStr
		}
		visited[ptr] = true
		defer delete(visited, ptr)

		elements := v.Value.(*AtomArray).Elements
		if len(elements) == 0 {
			return EmptyArrStr
		}

		// Use string builder for array formatting
		builder := StringBuilderPool.Get().(*strings.Builder)
		builder.Reset()
		builder.WriteByte('[')

		first := true
		for _, element := range elements {
			if !first {
				builder.WriteString(", ")
			}
			first = false
			builder.WriteString(valueToStringWithVisited(element, visited))
		}
		builder.WriteByte(']')
		result := builder.String()
		StringBuilderPool.Put(builder)
		return result

	case AtomTypeMethod:
		method := v.Value.(*AtomMethod)
		fnValue := method.Fn
		if CheckType(fnValue, AtomTypeFunc) {
			atomCode := fnValue.Value.(*AtomCode)
			return BuildString(func(b *strings.Builder) {
				b.WriteString("bound method ")
				b.WriteString(atomCode.Name)
				b.WriteByte('(')

				if atomCode.Argc == 0 {
					// No parameters
				} else {
					// Build parameters efficiently
					for i := 0; i < atomCode.Argc; i++ {
						if i > 0 {
							b.WriteString(", ")
						}
						b.WriteByte('$')
						b.WriteString(strconv.Itoa(i))
					}
				}
				b.WriteString("){}")
			})
		} else if CheckType(fnValue, AtomTypeNativeFunc) {
			nativeFunc := fnValue.Value.(NativeFunc)
			return BuildString(func(b *strings.Builder) {
				b.WriteString("bound method ")
				b.WriteString(nativeFunc.Name)
				b.WriteByte('(')

				switch nativeFunc.Paramc {
				case Variadict:
					b.WriteString("...")
				case 0:
					// No parameters
				default:
					// Build parameters efficiently
					for i := 0; i < nativeFunc.Paramc; i++ {
						if i > 0 {
							b.WriteString(", ")
						}
						b.WriteByte('$')
						b.WriteString(strconv.Itoa(i))
					}
				}
				b.WriteString("){}")
			})
		}
		return "bound method"

	case AtomTypeFunc:
		// Optimized: avoid fmt.Sprintf, use helper function
		atomCode := v.Value.(*AtomCode)
		return BuildString(func(b *strings.Builder) {
			b.WriteString("function ")
			b.WriteString(atomCode.Name)
			b.WriteByte('(')

			if atomCode.Argc == 0 {
				// No parameters
			} else {
				// Build parameters efficiently
				for i := 0; i < atomCode.Argc; i++ {
					if i > 0 {
						b.WriteString(", ")
					}
					b.WriteByte('$')
					b.WriteString(strconv.Itoa(i))
				}
			}
			b.WriteString("){}")
		})

	case AtomTypeNativeFunc:
		// Optimized: use helper function
		nativeFunc := v.Value.(NativeFunc)
		return BuildString(func(b *strings.Builder) {
			b.WriteString(nativeFunc.Name)
			b.WriteByte('(')

			switch nativeFunc.Paramc {
			case Variadict:
				b.WriteString("...")
			case 0:
				// No parameters
			default:
				// Build parameters efficiently
				for i := 0; i < nativeFunc.Paramc; i++ {
					if i > 0 {
						b.WriteString(", ")
					}
					b.WriteByte('$')
					b.WriteString(strconv.Itoa(i))
				}
			}
			b.WriteString("){}")
		})

	case AtomTypeErr:
		// Fast path: direct string conversion
		return v.Value.(string)

	case AtomTypePromise:
		promise := v.Value.(*AtomPromise)
		return BuildString(func(b *strings.Builder) {
			b.WriteString("Promise { ")
			if promise.State == PromiseStateFulfilled {
				if CheckType(promise.Value, AtomTypeStr) {
					b.WriteByte('\'')
					b.WriteString(promise.Value.String())
					b.WriteByte('\'')
				} else {
					b.WriteString(promise.Value.String())
				}
			} else {
				b.WriteString(promise.State.String())
			}
			b.WriteString(" }")
		})

	default:
		// Fallback to fmt.Sprintf only when necessary
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
		// Use a stable hash function for strings
		str := v.Value.(string)
		if len(str) == 0 {
			return 0
		}
		hash := 5381
		for i := 0; i < len(str); i++ {
			hash = ((hash << 5) + hash) + int(str[i])
		}
		return hash

	case AtomTypeNull:
		return 0

	case AtomTypeClass:
		atomClass := v.Value.(*AtomClass)
		return int(uintptr(unsafe.Pointer(atomClass)))

	case AtomTypeClassInstance:
		instance := v.Value.(*AtomClassInstance)
		return int(uintptr(unsafe.Pointer(instance)))

	case AtomTypeEnum:
		return v.Value.(*AtomObject).HashValue()

	case AtomTypeObj:
		return v.Value.(*AtomObject).HashValue()

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

	case AtomTypePromise:
		promise := v.Value.(*AtomPromise)
		return promise.Value.HashValue()

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
	case AtomTypeClass:
		return "class"
	case AtomTypeClassInstance:
		instance := value.Value.(*AtomClassInstance)
		return instance.Prototype.Value.(*AtomClass).Name
	case AtomTypeEnum:
		return "enum"
	case AtomTypeObj:
		return "object"
	case AtomTypeArray:
		return "array"
	case AtomTypeMethod:
		return "method"
	case AtomTypeFunc:
		return "function"
	case AtomTypeNativeFunc:
		return "native function"
	case AtomTypeErr:
		return "error"
	default:
		return fmt.Sprintf("unknown type: %d", value.Type)
	}
}

// Helper function for optimized string building with pooling
func BuildString(fn func(*strings.Builder)) string {
	builder := StringBuilderPool.Get().(*strings.Builder)
	builder.Reset()
	fn(builder)
	result := builder.String()
	StringBuilderPool.Put(builder)
	return result
}

// Helper function to get string representation of a value, optimized for string types
func ValueToString(v *AtomValue) string {
	if CheckType(v, AtomTypeStr) {
		// Fast path: add quotes directly
		return "'" + v.Value.(string) + "'"
	}
	return v.String()
}

// Helper function to get string representation with visited tracking for circular reference detection
func valueToStringWithVisited(v *AtomValue, visited map[uintptr]bool) string {
	if CheckType(v, AtomTypeStr) {
		// Fast path: add quotes directly
		return "'" + v.Value.(string) + "'"
	}
	return v.stringWithVisited(visited)
}
