package runtime

func IsNumberType(value *AtomValue) bool {
	return CheckType(value, AtomTypeInt) || CheckType(value, AtomTypeNum) || CheckType(value, AtomTypeBigInt)
}

func IsInteger(value float64) bool {
	return value == float64(int64(value))
}

func CheckType(value *AtomValue, ttype AtomType) bool {
	return value.Type == ttype
}

func CheckTypeAsync(value *AtomValue) bool {
	return CheckType(value, AtomTypeFunc) && value.Obj.(*AtomCode).Async
}
