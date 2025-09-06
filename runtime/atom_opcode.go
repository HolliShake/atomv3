package runtime

type OpCode byte

const (
	OpLoadInt OpCode = iota
	OpLoadNum
	OpLoadStr
	OpLoadBool
	OpLoadNull
	OpLoadLocal    // with 4 bytes argument
	OpLoadFunction // with 4 bytes argument
	OpCall         // with 4 bytes argument
	OpMul
	OpDiv
	OpMod
	OpAdd
	OpSub
	OpShl
	OpShr
	OpStoreLocal
	OpJumpIfTrueOrPop
	OpPopJumpIfFalse
	OpJump
	OpNoOp
	OpPopTop
	OpReturn
)
