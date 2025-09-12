package runtime

type OpCode byte

const (
	OpLoadInt OpCode = iota
	OpLoadNum
	OpLoadStr // With N bytes argument
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
	OpPluckAttribute // with N bytes argument
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
	OpJumpIfFalseOrPop  // with 4 bytes argument a.k.a jump offset
	OpJumpIfTrueOrPop   // with 4 bytes argument a.k.a jump offset
	OpPopJumpIfFalse    // with 4 bytes argument a.k.a jump offset
	OpPopJumpIfTrue     // with 4 bytes argument a.k.a jump offset
	OpPeekJumpIfEqual   // with 4 bytes argument a.k.a jump offset
	OpPopJumpIfNotError // with 4 bytes argument a.k.a jump offset
	OpJump              // with 4 bytes argument a.k.a jump offset
	OpAbsoluteJump      // with 4 bytes argument a.k.a jump offset
	OpDupTop
	OpNoOp
	OpPopTop
	OpReturn
)
