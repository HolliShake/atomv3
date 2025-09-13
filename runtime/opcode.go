package runtime

type OpCode byte

const (
	OpLoadInt           OpCode = iota
	OpLoadNum                  //
	OpLoadStr                  // With N bytes argument
	OpLoadBool                 //
	OpLoadNull                 //
	OpLoadArray                // with 4 bytes argument
	OpLoadObject               // with 4 bytes argument
	OpLoadName                 // with 4 bytes argument
	OpLoadModule0              // With N bytes argument // name -> builtin
	OpLoadModule1              // With N bytes argument // path
	OpLoadFunction             // with 4 bytes argument
	OpMakeClass                // with 4 bytes argument
	OpExtendClass              //
	OpMakeEnum                 // with 4 bytes argument
	OpCall                     // with 4 bytes argument
	OpNot                      //
	OpNeg                      //
	OpPos                      //
	OpIndex                    //
	OpPluckAttribute           // with N bytes argument
	OpMul                      //
	OpDiv                      //
	OpMod                      //
	OpAdd                      //
	OpSub                      //
	OpShl                      //
	OpShr                      //
	OpCmpLt                    //
	OpCmpLte                   //
	OpCmpGt                    //
	OpCmpGte                   //
	OpCmpEq                    //
	OpCmpNe                    //
	OpAnd                      //
	OpOr                       //
	OpXor                      //
	OpInitVar                  // with (N + 1 + 1) bytes argument
	OpStoreFast                // with N bytes argument
	OpStoreLocal               // with N bytes argument
	OpSetIndex                 //
	OpJumpIfFalseOrPop         // with 4 bytes argument a.k.a jump offset
	OpJumpIfTrueOrPop          // with 4 bytes argument a.k.a jump offset
	OpPopJumpIfFalse           // with 4 bytes argument a.k.a jump offset
	OpPopJumpIfTrue            // with 4 bytes argument a.k.a jump offset
	OpPeekJumpIfEqual          // with 4 bytes argument a.k.a jump offset
	OpPopJumpIfNotError        // with 4 bytes argument a.k.a jump offset
	OpJump                     // with 4 bytes argument a.k.a jump offset
	OpAbsoluteJump             // with 4 bytes argument a.k.a jump offset
	OpEnterBlock               //
	OpExitBlock                //
	OpDupTop                   //
	OpNoOp                     //
	OpPopTop                   //
	OpReturn                   //
)
