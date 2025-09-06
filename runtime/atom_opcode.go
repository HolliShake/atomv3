package runtime

type OpCode byte

const (
	OpLoadInt OpCode = iota
	OpLoadNum
	OpLoadStr
	OpLoadBool
	OpLoadNull
	OpLoadGlobal   // with 4 bytes argument
	OpLoadLocal    // with 4 bytes argument
	OpLoadCapture  // with 4 bytes argument
	OpLoadFunction // with 4 bytes argument
	OpCall         // with 4 bytes argument
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
	OpStoreLocal
	OpStoreCapture
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
