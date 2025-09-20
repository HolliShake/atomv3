package runtime

type OpCode byte

const (
	OpMakeModule OpCode = iota + 69 // with 4 bytes argument
	OpLoadInt
	OpLoadNum           //
	OpLoadStr           // with N bytes argument
	OpLoadBool          //
	OpLoadNull          //
	OpLoadArray         // with 4 bytes argument
	OpLoadObject        // with 4 bytes argument
	OpLoadName          // with 4 bytes argument
	OpLoadCapture       // with 4 bytes argument
	OpLoadModule        // With N bytes argument
	OpLoadFunction      // with 4 bytes argument
	OpMakeClass         // with 4 bytes argument
	OpExtendClass       //
	OpMakeEnum          // with 4 bytes argument
	OpCallConstructor   // with 4 bytes argument
	OpCall              // with 4 bytes argument
	OpAwait             //
	OpNot               //
	OpNeg               //
	OpPos               //
	OpTypeof            //
	OpIndex             //
	OpPluckAttribute    // with N bytes argument
	OpMul               //
	OpDiv               //
	OpMod               //
	OpAdd               //
	OpSub               //
	OpShl               //
	OpShr               //
	OpCmpLt             //
	OpCmpLte            //
	OpCmpGt             //
	OpCmpGte            //
	OpCmpEq             //
	OpCmpNe             //
	OpAnd               //
	OpOr                //
	OpXor               //
	OpStoreModule       // with N bytes argument
	OpStoreCapture      // with 4 bytes argument
	OpStoreLocal        // with 4 bytes argument
	OpSetIndex          //
	OpJumpIfFalseOrPop  // with 4 bytes argument a.k.a jump offset
	OpJumpIfTrueOrPop   // with 4 bytes argument a.k.a jump offset
	OpPopJumpIfFalse    // with 4 bytes argument a.k.a jump offset
	OpPopJumpIfTrue     // with 4 bytes argument a.k.a jump offset
	OpPeekJumpIfEqual   // with 4 bytes argument a.k.a jump offset
	OpPopJumpIfNotError // with 4 bytes argument a.k.a jump offset
	OpJump              // with 4 bytes argument a.k.a jump offset
	OpAbsoluteJump      // with 4 bytes argument a.k.a jump offset
	OpDupTop            //
	OpNoOp              //
	OpPopTop            //
	OpReturn            //
	// max 255
)
