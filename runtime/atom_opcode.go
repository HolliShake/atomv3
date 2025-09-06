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
	OpLoadCapture  // with 4 bytes argument
	OpLoadFunction // with 4 bytes argument
	OpCall         // with 4 bytes argument
	OpIndex
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
	OpStoreGlobal  // with 4 bytes argument | alias for OpStoreLocal
	OpStoreCapture // with 4 bytes argument
	OpStoreLocal   // with 4 bytes argument
	OpJumpIfFalseOrPop
	OpJumpIfTrueOrPop
	OpPopJumpIfFalse
	OpPopJumpIfTrue
	OpJump
	OpDupTop
	OpNoOp
	OpPopTop
	OpReturn
)
