package runtime

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
)

func ReadInt(data []OpCode, offset int) int {
	return int(binary.LittleEndian.Uint32([]byte{byte(data[offset]), byte(data[offset+1]), byte(data[offset+2]), byte(data[offset+3])}))
}

func ReadNum(data []OpCode, offset int) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64([]byte{byte(data[offset]), byte(data[offset+1]), byte(data[offset+2]), byte(data[offset+3]), byte(data[offset+4]), byte(data[offset+5]), byte(data[offset+6]), byte(data[offset+7])}))
}

func ReadStr(data []OpCode, offset int) string {
	bytes := []byte{}
	for i := offset; i < len(data) && data[i] != 0; i++ {
		bytes = append(bytes, byte(data[i]))
	}
	return string(bytes)
}

func CoerceToInt(value *AtomValue) int32 {
	switch value.Type {
	case AtomTypeInt:
		return value.I32
	case AtomTypeNum:
		return int32(value.F64)
	case AtomTypeBigInt:
		bigVal := value.Obj.(*big.Int)
		if !bigVal.IsInt64() {
			return 0
		}
		val := bigVal.Int64()
		if val > math.MaxInt32 || val < math.MinInt32 {
			return 0
		}
		return int32(val)
	case AtomTypeStr:
		val, err := strconv.ParseInt(value.Str, 10, 32)
		if err != nil {
			return 0
		}
		return int32(val)
	case AtomTypeBool:
		if value.I32 == 1 {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func CoerceToLong(value *AtomValue) int64 {
	switch value.Type {
	case AtomTypeInt:
		return int64(value.I32)
	case AtomTypeNum:
		return int64(value.F64)
	case AtomTypeBigInt:
		bigVal := value.Obj.(*big.Int)
		if !bigVal.IsInt64() {
			return 0
		}
		val := bigVal.Int64()
		return val
	case AtomTypeStr:
		val, err := strconv.ParseInt(value.Str, 10, 64)
		if err != nil {
			return 0
		}
		return val
	case AtomTypeBool:
		if value.I32 == 1 {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func CoerceToNum(value *AtomValue) float64 {
	switch value.Type {
	case AtomTypeInt:
		return float64(value.I32)
	case AtomTypeNum:
		return value.F64
	case AtomTypeBigInt:
		bigVal := value.Obj.(*big.Int)
		if !bigVal.IsInt64() {
			return 0
		}
		val := bigVal.Int64()
		return float64(val)
	case AtomTypeStr:
		val, err := strconv.ParseFloat(value.Str, 64)
		if err != nil {
			return 0
		}
		return val
	case AtomTypeBool:
		if value.I32 == 1 {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func CoerceToBigInt(value *AtomValue) *big.Int {
	switch value.Type {
	case AtomTypeInt:
		return BigInt(strconv.FormatInt(int64(value.I32), 10))
	case AtomTypeNum:
		return BigInt(strconv.FormatFloat(value.F64, 'g', -1, 64))
	case AtomTypeBigInt:
		return value.Obj.(*big.Int)
	case AtomTypeStr:
		return BigInt(value.Str)
	case AtomTypeBool:
		if value.I32 == 1 {
			return BigInt("1")
		}
		return BigInt("0")
	default:
		return BigInt("0")
	}
}

func CoerceToBool(value *AtomValue) bool {
	switch value.Type {
	case AtomTypeInt:
		return value.I32 != 0
	case AtomTypeNum:
		return value.F64 != 0
	case AtomTypeBigInt:
		return value.Obj.(*big.Int).Sign() != 0
	case AtomTypeStr:
		return value.Str != ""
	case AtomTypeBool:
		return value.I32 == 1
	case AtomTypeNull:
		return false
	default:
		return true
	}
}

func FormatError(frame *AtomCallFrame, message string) string {
	file := frame.Fn.Obj.(*AtomCode).File

	ip := frame.Ip

	// binary search the line
	line := BinarySearch(frame.Fn.Obj.(*AtomCode).Line, ip)

	return fmt.Sprintf("[%s:%d]::Error: %s", file, line, message)
}

func BinarySearch(lines []AtomDebugLine, ip int) int {
	left, right := 0, len(lines)-1
	result := -1

	for left <= right {
		mid := (left + right) / 2
		if lines[mid].Address <= ip {
			result = lines[mid].Line
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return result
}

func BigInt(v string) *big.Int {
	// Set precision to a very high value to simulate Python's arbitrary precision
	// Python's int has unlimited precision, so we use a very high precision for big.Float
	val, ok := big.NewInt(0).SetString(v, 10)
	if !ok {
		return big.NewInt(0)
	}
	return val
}

func CleanupStack(frame *AtomCallFrame, count int) {
	frame.Stack.PopN(count)
}

func SerializeObject(obj *AtomValue) any /* native go object only */ {
	seen := make(map[*AtomValue]bool)
	return serializeObjectHelper(obj, seen)
}

func serializeObjectHelper(obj *AtomValue, seen map[*AtomValue]bool) any {
	// Check for circular reference
	if obj.Type == AtomTypeArray || obj.Type == AtomTypeObj || obj.Type == AtomTypeClassInstance {
		if seen[obj] {
			return nil // or return a placeholder like "<circular>"
		}
		seen[obj] = true
	}

	switch obj.Type {
	case AtomTypeNull:
		return nil
	case AtomTypeInt:
		return obj.I32
	case AtomTypeNum:
		return obj.F64
	case AtomTypeBool:
		return obj.I32 == 1
	case AtomTypeBigInt:
		return obj.Obj.(*big.Int).String()
	case AtomTypeStr:
		return obj.Str
	case AtomTypeArray:
		arr := obj.Obj.(*AtomArray)
		result := make([]any, len(arr.Elements))
		for i, elem := range arr.Elements {
			result[i] = serializeObjectHelper(elem, seen)
		}
		return result
	case AtomTypeObj:
		m := obj.Obj.(*AtomObject)
		result := make(map[string]any)
		for k, v := range m.Elements {
			result[k] = serializeObjectHelper(v, seen)
		}
		return result
	case AtomTypeClass:
		atomClass := obj.Obj.(*AtomClass)
		return map[string]any{
			"__class__": atomClass.Name,
		}
	case AtomTypeClassInstance:
		classInstance := obj.Obj.(*AtomClassInstance)
		prototype := classInstance.Prototype.Obj.(*AtomClass)

		// Check if the property is an AtomObject (not a builtin)
		if reflect.TypeOf(classInstance.Property.Obj) != reflect.TypeOf(&AtomValue{}) {
			// For builtin class instances, return class name only
			return map[string]any{
				"__class__": prototype.Name,
			}
		}

		properties := classInstance.Property.Obj.(*AtomObject).Elements
		result := make(map[string]any)
		result["__class__"] = prototype.Name

		for k, v := range properties {
			result[k] = serializeObjectHelper(v, seen)
		}
		return result
	default:
		return nil
	}
}

func ToAtomObject(data any) *AtomValue {
	switch val := data.(type) {
	case int:
		{
			return NewAtomValueInt(val)
		}
	case float64:
		{
			return NewAtomValueNum(val)
		}
	case string:
		{
			return NewAtomValueStr(val)
		}
	case bool:
		{
			if val {
				return NewAtomValueInt(1)
			} else {
				return NewAtomValueInt(0)
			}
		}
	case map[string]any:
		{
			result := map[string]*AtomValue{}
			for k, v := range val {
				result[k] = ToAtomObject(v)
			}
			return NewAtomGenericValue(AtomTypeObj, NewAtomObject(result))
		}
	case []any:
		{
			arr := val
			atomArr := &AtomArray{Elements: make([]*AtomValue, len(arr))}
			for i, elem := range arr {
				atomArr.Elements[i] = ToAtomObject(elem)
			}
			return NewAtomGenericValue(AtomTypeArray, atomArr)
		}
	case *big.Int:
		{
			return NewAtomValueBigInt(val)
		}
	case *big.Float:
		{
			return NewAtomValueNum(0) // Placeholder since NewAtomValueBigFloat doesn't exist
		}
	case *AtomValue:
		{
			return val
		}
	default:
		return NewAtomValueStr("null")
	}
}
