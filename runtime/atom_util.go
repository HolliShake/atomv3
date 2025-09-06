package runtime

import (
	"encoding/binary"
	"math"
	"strconv"
	"unsafe"
)

func ReadInt(data []OpCode, offset int) int {
	return int(binary.LittleEndian.Uint32([]byte{byte(data[offset]), byte(data[offset+1]), byte(data[offset+2]), byte(data[offset+3])}))
}

func ReadNum(data []OpCode, offset int) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64([]byte{byte(data[offset]), byte(data[offset+1]), byte(data[offset+2]), byte(data[offset+3]), byte(data[offset+4]), byte(data[offset+5]), byte(data[offset+6]), byte(data[offset+7])}))
}

func ReadStr(data []OpCode, offset int) string {
	bytes := make([]byte, 0)
	for i := offset; data[i] != 0; i++ {
		bytes = append(bytes, byte(data[i]))
	}
	return string(bytes)
}

func CoerceToInt(value *AtomValue) int32 {
	switch value.Type {
	case AtomTypeInt:
		return value.Value.(int32)
	case AtomTypeNum:
		return int32(value.Value.(float64))
	case AtomTypeStr:
		val, err := strconv.Atoi(value.Value.(string))
		if err != nil {
			return 0
		}
		return int32(val)
	case AtomTypeBool:
		if value.Value.(bool) {
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
		return int64(value.Value.(int32))
	case AtomTypeNum:
		return int64(value.Value.(float64))
	case AtomTypeStr:
		val, err := strconv.ParseInt(value.Value.(string), 10, 64)
		if err != nil {
			return 0
		}
		return val
	case AtomTypeBool:
		if value.Value.(bool) {
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
		return float64(value.Value.(int32))
	case AtomTypeNum:
		return value.Value.(float64)
	case AtomTypeStr:
		val, err := strconv.ParseFloat(value.Value.(string), 64)
		if err != nil {
			return 0
		}
		return val
	case AtomTypeBool:
		if value.Value.(bool) {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func CoerceToBool(value *AtomValue) bool {
	switch value.Type {
	case AtomTypeInt:
		return value.Value.(int32) != 0
	case AtomTypeNum:
		return value.Value.(float64) != 0
	case AtomTypeStr:
		return value.Value.(string) != ""
	case AtomTypeBool:
		return value.Value.(bool)
	case AtomTypeNull:
		return false
	default:
		return true
	}
}

func PointerToUintPointer(cell *AtomCell) uintptr {
	return uintptr(unsafe.Pointer(cell))
}

func UintPointerToPointer(value uintptr) *AtomCell {
	return (*AtomCell)(unsafe.Pointer(value))
}
