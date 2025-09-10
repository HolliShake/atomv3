package runtime

type OpCode byte

const (
	OpLoadInt OpCode = iota
	OpLoadNum
	OpLoadStr
	OpLoadBool
	OpLoadNull
	OpLoadArray    // with 4 bytes argument
	OpLoadObject   // with 4 bytes argument
	OpLoadLocal    // with 4 bytes argument
	OpLoadModule0  // With N bytes argument // name -> builtin
	OpLoadModule1  // With N bytes argument // path
	OpLoadFunction // with 4 bytes argument
	OpMakeEnum
	OpCall // with 4 bytes argument
	OpNot
	OpNeg
	OpPos
	OpIndex
	OpPluckAttribute
	OpMul
	OpDiv
	OpMod
	OpAdd
	OpSub
	OpShl
	OpShr
	OpCmpLt
	OpCmpLte
	OpCmpGt
	OpCmpGte
	OpCmpEq
	OpCmpNe
	OpAnd
	OpOr
	OpXor
	OpStoreGlobal // with 4 bytes argument | alias for OpStoreLocal
	OpStoreLocal  // with 4 bytes argument
	OpSetIndex
	OpJumpIfFalseOrPop
	OpJumpIfTrueOrPop
	OpPopJumpIfFalse
	OpPopJumpIfTrue
	OpPeekJumpIfEqual
	OpPopJumpIfNotError
	OpJump
	OpAbsoluteJump
	OpDupTop
	OpNoOp
	OpPopTop
	OpReturn
)
