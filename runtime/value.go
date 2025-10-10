package runtime

import (
	"fmt"
	"math"
	"math/big"
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
	AtomTypeBigInt
	AtomTypeBool
	AtomTypeStr
	AtomTypeNull
	AtomTypeClass
	AtomTypeClassInstance
	AtomTypeEnum
	AtomTypeObj
	AtomTypeArray
	AtomTypeMethod
	AtomTypeNativeMethod
	AtomTypeFunc
	AtomTypeNativeFunc
	AtomTypeErr
	AtomTypePromise
	AtomTypeExternal
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
	Type AtomType
	I32  int32
	F64  float64
	Str  string
	Obj  any // only for complex cases
}

type NativeFunction func(intereter *AtomInterpreter, argc int)

func NewAtomValue(atomType AtomType) *AtomValue {
	obj := &AtomValue{}
	obj.Type = atomType
	return obj
}

func NewAtomValueInt(value int) *AtomValue {
	obj := NewAtomValue(AtomTypeInt)
	obj.I32 = int32(value)
	return obj
}

func NewAtomValueNum(value float64) *AtomValue {
	obj := NewAtomValue(AtomTypeNum)
	obj.F64 = value
	return obj
}

func NewAtomValueBigInt(value *big.Int) *AtomValue {
	obj := NewAtomValue(AtomTypeBigInt)
	obj.Obj = value
	return obj
}

func NewAtomValueFalse() *AtomValue {
	obj := NewAtomValue(AtomTypeBool)
	obj.I32 = 0
	return obj
}

func NewAtomValueTrue() *AtomValue {
	obj := NewAtomValue(AtomTypeBool)
	obj.I32 = 1
	return obj
}

func NewAtomValueStr(value string) *AtomValue {
	obj := NewAtomValue(AtomTypeStr)
	obj.Str = value
	return obj
}

func NewAtomValueNull() *AtomValue {
	obj := NewAtomValue(AtomTypeNull)
	obj.I32 = 0
	return obj
}

func NewAtomValueError(message string) *AtomValue {
	obj := NewAtomValue(AtomTypeErr)
	obj.Str = message
	return obj
}

func NewAtomGenericValue(atomType AtomType, value any) *AtomValue {
	obj := NewAtomValue(atomType)
	obj.Obj = value
	return obj
}

func (v *AtomValue) String() string {
	return v.StringWithVisited(make(map[uintptr]bool))
}

func (v *AtomValue) StringWithVisited(visited map[uintptr]bool) string {
	switch v.Type {
	case AtomTypeInt:
		// Fast path: direct conversion without fmt.Sprintf
		return strconv.FormatInt(int64(v.I32), 10)

	case AtomTypeNum:
		// Fast path: direct conversion without fmt.Sprintf
		return strconv.FormatFloat(v.F64, 'f', -1, 64)

	case AtomTypeBigInt:
		// Fast path: direct conversion without fmt.Sprintf
		return v.Obj.(*big.Int).Text(10)

	case AtomTypeBool:
		// Fast path: pre-allocated strings
		if v.I32 == 1 {
			return TrueStr
		}
		return FalseStr

	case AtomTypeStr:
		// Fast path: direct return
		return v.Str

	case AtomTypeNull:
		// Fast path: pre-allocated string
		return NullStr

	case AtomTypeClass:
		// Optimized: avoid fmt.Sprintf, use helper function
		atomClass := v.Obj.(*AtomClass)
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

		classInstance := v.Obj.(*AtomClassInstance)

		// For external type | builtin classess
		if classInstance.Property.Type == AtomTypeExternal {
			// For builtin class, obj might not be an AtomObject
			builder := StringBuilderPool.Get().(*strings.Builder)
			builder.Reset()
			builder.WriteString(classInstance.Prototype.Obj.(*AtomClass).Name)
			builder.WriteString(EmptyClassInstanceStr)
			return builder.String()
		}

		prototype := classInstance.Prototype.Obj.(*AtomClass)
		properties := classInstance.Property.Obj.(*AtomObject).Elements

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
			builder.WriteString(ValueToStringWithVisited(value, visited))
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

		enumElements := v.Obj.(*AtomObject).Elements
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
			builder.WriteString(ValueToStringWithVisited(value, visited))
		}
		builder.WriteByte('}')
		result := builder.String()
		StringBuilderPool.Put(builder)
		return result

	case AtomTypeObj:
		// Check for self-reference
		ptr := uintptr(unsafe.Pointer(v))
		if visited[ptr] {
			return "SelfRefStr"
		}
		visited[ptr] = true
		defer delete(visited, ptr)

		objElements := v.Obj.(*AtomObject).Elements
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
			builder.WriteString(ValueToStringWithVisited(value, visited))
		}
		builder.WriteByte('}')
		result := builder.String()
		StringBuilderPool.Put(builder)
		return result

	case AtomTypeArray:
		// Check for self-reference
		ptr := uintptr(unsafe.Pointer(v))
		if visited[ptr] {
			return "[[self]]"
		}
		visited[ptr] = true
		defer delete(visited, ptr)

		elements := v.Obj.(*AtomArray).Elements
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
			builder.WriteString(ValueToStringWithVisited(element, visited))
		}
		builder.WriteByte(']')
		result := builder.String()
		StringBuilderPool.Put(builder)
		return result

	case AtomTypeMethod:
		method := v.Obj.(*AtomMethod)
		fnValue := method.Fn
		if CheckType(fnValue, AtomTypeFunc) {
			atomCode := fnValue.Obj.(*AtomCode)
			return BuildString(func(b *strings.Builder) {
				b.WriteString("<bound method ")
				b.WriteString(GetTypeString(method.This))
				b.WriteByte('.')
				b.WriteString(atomCode.Name)
				b.WriteString(" of ")
				b.WriteString(method.This.String())
				b.WriteByte('>')
			})
		} else if CheckType(fnValue, AtomTypeNativeFunc) {
			nativeFunc := fnValue.Obj.(*AtomNativeFunc)
			return BuildString(func(b *strings.Builder) {
				b.WriteString("<bound method ")
				b.WriteString(GetTypeString(method.This))
				b.WriteByte('.')
				b.WriteString(nativeFunc.Name)
				b.WriteString(" of ")
				b.WriteString(method.This.String())
				b.WriteByte('>')
			})
		}
		return "<bound method>"

	case AtomTypeNativeMethod:
		nativeMethod := v.Obj.(*AtomNativeMethod)
		return BuildString(func(b *strings.Builder) {
			b.WriteString("<bound method ")
			b.WriteString(nativeMethod.Name)
			b.WriteString(" of ")
			b.WriteString(nativeMethod.This.String())
			b.WriteByte('>')
		})

	case AtomTypeFunc:
		// Optimized: avoid fmt.Sprintf, use helper function
		atomCode := v.Obj.(*AtomCode)
		return BuildString(func(b *strings.Builder) {
			b.WriteString("<function ")
			b.WriteString(atomCode.Name)
			b.WriteString(" at ")
			b.WriteString(atomCode.File)
			b.WriteByte('>')
		})

	case AtomTypeNativeFunc:
		// Optimized: use helper function
		nativeFunc := v.Obj.(*AtomNativeFunc)
		return BuildString(func(b *strings.Builder) {
			b.WriteString("<built-in function ")
			b.WriteString(nativeFunc.Name)
			b.WriteByte('>')
		})

	case AtomTypeErr:
		// Fast path: direct string conversion
		return v.Str

	case AtomTypePromise:
		promise := v.Obj.(*AtomPromise)
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
		return fmt.Sprintf("%v", v.Obj)
	}
}

func (v *AtomValue) HashValue() int {
	switch v.Type {
	case AtomTypeInt:
		return int(v.I32)

	case AtomTypeNum:
		// Use math.Float64bits for consistent hashing of float64
		bits := math.Float64bits(v.F64)
		return int(bits ^ (bits >> 32))

	case AtomTypeBigInt:
		str := v.Obj.(*big.Int).String()
		hash := 0
		for _, b := range []byte(str) {
			hash = ((hash << 5) + hash) + int(b)
		}
		return hash

	case AtomTypeBool:
		// Branchless boolean to int conversion
		return int(*(*uint8)(unsafe.Pointer(&v.I32)))

	case AtomTypeStr:
		// Use a stable hash function for strings
		str := v.Str
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
		atomClass := v.Obj.(*AtomClass)
		return int(uintptr(unsafe.Pointer(atomClass)))

	case AtomTypeClassInstance:
		instance := v.Obj.(*AtomClassInstance)
		return int(uintptr(unsafe.Pointer(instance)))

	case AtomTypeEnum:
		return v.Obj.(*AtomObject).HashValue()

	case AtomTypeObj:
		return v.Obj.(*AtomObject).HashValue()

	case AtomTypeArray:
		elements := v.Obj.(*AtomArray).Elements
		if len(elements) == 0 {
			return 0
		}
		hash := 0
		for _, element := range elements {
			hash = hash*31 + element.HashValue()
		}
		return hash

	case AtomTypeMethod:
		nfn := v.Obj.(*AtomMethod).Fn
		return nfn.Obj.(*AtomCode).HashValue()

	case AtomTypeNativeMethod:
		fn := v.Obj.(*AtomNativeMethod)
		return int(reflect.ValueOf(fn.Callable).Pointer())

	case AtomTypeFunc:
		return v.Obj.(*AtomCode).HashValue()

	case AtomTypeNativeFunc:
		f := v.Obj.(AtomNativeFunc)
		return int(reflect.ValueOf(f.Callable).Pointer())

	case AtomTypeErr:
		// Use unsafe string to bytes conversion to avoid allocation
		str := v.Str
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
		promise := v.Obj.(*AtomPromise)
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
	case AtomTypeBigInt:
		return "big number"
	case AtomTypeBool:
		return "bool"
	case AtomTypeStr:
		return "string"
	case AtomTypeNull:
		return "null"
	case AtomTypeClass:
		return "class"
	case AtomTypeClassInstance:
		instance := value.Obj.(*AtomClassInstance)
		return instance.Prototype.Obj.(*AtomClass).Name
	case AtomTypeEnum:
		return "enum"
	case AtomTypeObj:
		return "object"
	case AtomTypeArray:
		return "array"
	case AtomTypeMethod:
		return "method"
	case AtomTypeNativeMethod:
		return "native method"
	case AtomTypeFunc:
		return "function"
	case AtomTypeNativeFunc:
		return "native function"
	case AtomTypeErr:
		return "error"
	case AtomTypePromise:
		return "promise"
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
		return "'" + v.Str + "'"
	}
	return v.String()
}

// Helper function to get string representation with visited tracking for circular reference detection
func ValueToStringWithVisited(v *AtomValue, visited map[uintptr]bool) string {
	if CheckType(v, AtomTypeStr) {
		// Fast path: add quotes directly
		return "'" + v.Str + "'"
	}
	return v.StringWithVisited(visited)
}
