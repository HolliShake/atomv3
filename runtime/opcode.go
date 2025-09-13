package runtime

type OpCode byte

const (
	OpLoadInt OpCode = iota
	OpLoadNum
	OpLoadStr      // With N bytes argument
	OpLoadBool     //
	OpLoadNull     //
	OpLoadArray    // with 4 bytes argument
	OpLoadObject   // with 4 bytes argument
	OpLoadLocal    // with 4 bytes argument
	OpLoadModule0  // With N bytes argument // name -> builtin
	OpLoadModule1  // With N bytes argument // path
	OpLoadFunction // with 4 bytes argument
	OpMakeClass    // with 4 bytes argument
	OpExtendClass  //
	OpMakeEnum     // with 4 bytes argument
	OpCall         // with 4 bytes argument
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
	OpInitVar           // with (N + 1 + 1) bytes argument
	OpStoreFast         // with N bytes argument
	OpStoreLocal        // with N bytes argument
	OpSetIndex          //
	OpJumpIfFalseOrPop  // with N bytes argument a.k.a jump offset
	OpJumpIfTrueOrPop   // with N bytes argument a.k.a jump offset
	OpPopJumpIfFalse    // with N bytes argument a.k.a jump offset
	OpPopJumpIfTrue     // with N bytes argument a.k.a jump offset
	OpPeekJumpIfEqual   // with N bytes argument a.k.a jump offset
	OpPopJumpIfNotError // with N bytes argument a.k.a jump offset
	OpJump              // with N bytes argument a.k.a jump offset
	OpAbsoluteJump      // with N bytes argument a.k.a jump offset
	OpDupTop
	OpNoOp
	OpRot2
	OpPopTop
	OpReturn
)
