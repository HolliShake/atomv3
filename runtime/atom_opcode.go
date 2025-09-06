package runtime

type OpCode byte

const (
	OpLoadInt OpCode = iota
	OpLoadNum
	OpLoadStr
	OpLoadBool
	OpMul
	OpDiv
	OpMod
	OpAdd
	OpSub
	OpStoreLocal
	OpJumpIfTrueOrPop
	OpReturn
)
