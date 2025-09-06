package runtime

import "fmt"

const (
	TRESHOLD = 1000
)

type AtomInterpreter struct {
	state           *AtomState
	Frame           *AtomStack
	EvaluationStack *AtomStack
	GcRoot          *AtomValue
	Allocation      int
}

func NewInterpreter(state *AtomState) *AtomInterpreter {
	return &AtomInterpreter{
		state:           state,
		Frame:           NewAtomStack(),
		EvaluationStack: NewAtomStack(),
		GcRoot:          NewAtomValue(AtomTypeObj),
	}
}

func (i *AtomInterpreter) pushVal(value *AtomValue) {
	i.GcRoot.Next = value
	i.GcRoot = value
	i.EvaluationStack.Push(value)
	i.Allocation++
}

func (i *AtomInterpreter) pushRef(value *AtomValue) {
	i.EvaluationStack.Push(value)
}

func (i *AtomInterpreter) pop() *AtomValue {
	return i.EvaluationStack.Pop()
}

func (i *AtomInterpreter) peek() *AtomValue {
	return i.EvaluationStack.Peek()
}

func (i *AtomInterpreter) executeFrame(parent *AtomValue, frame *AtomValue, offset int) {
	// Frame here is a function

	offsetStart := offset

	code := frame.Value.(*AtomCode)
	size := len(code.Code)

	forward := func(offset int) {
		offsetStart += offset
	}

	jump := func(offset int) {
		offsetStart = offset
	}

	for offsetStart < size {
		opCode := code.Code[offsetStart]
		offsetStart++

		if (i.Allocation % TRESHOLD) == 0 {
			// TODO: Garbage collection
		}

		switch opCode {
		case OpLoadInt:
			value := ReadInt(code.Code, offsetStart)
			i.pushVal(
				NewAtomValueInt(value),
			)
			forward(4)

		case OpLoadNum:
			value := ReadNum(code.Code, offsetStart)
			i.pushVal(
				NewAtomValueNum(value),
			)
			forward(8)

		case OpLoadStr:
			value := ReadStr(code.Code, offsetStart)
			i.pushVal(
				NewAtomValueStr(value),
			)
			forward(len(value) + 1)

		case OpLoadNull:
			i.pushRef(
				i.state.NullValue,
			)

		case OpLoadArray:
			length := ReadInt(code.Code, offsetStart)
			elements := make([]*AtomValue, 0)
			for range length {
				elements = append(elements, i.pop())
			}
			i.pushVal(
				NewAtomValueArray(elements),
			)
			forward(4)

		case OpLoadFunction:
			offset := ReadInt(code.Code, offsetStart)
			fn := i.state.FunctionTable.Get(offset)
			i.pushRef(fn)
			forward(4)

		case OpCall:
			argc := ReadInt(code.Code, offsetStart)
			forward(4)
			call := i.pop()
			global := parent
			if global == nil {
				global = frame
			}
			DoCall(i, global, call, argc)

		case OpLoadCapture:
			index := ReadInt(code.Code, offsetStart)
			value := code.Env1[index]
			i.pushRef(value)
			forward(4)

		case OpLoadLocal:
			index := ReadInt(code.Code, offsetStart)
			value := code.Env0[index]
			i.pushRef(value)
			forward(4)

		case OpMul:
			rhs := i.pop()
			lhs := i.pop()
			DoMultiplication(i, lhs, rhs)

		case OpDiv:
			rhs := i.pop()
			lhs := i.pop()
			DoDivision(i, lhs, rhs)

		case OpMod:
			rhs := i.pop()
			lhs := i.pop()
			DoModulus(i, lhs, rhs)

		case OpAdd:
			rhs := i.pop()
			lhs := i.pop()
			DoAddition(i, lhs, rhs)

		case OpSub:
			rhs := i.pop()
			lhs := i.pop()
			DoSubtraction(i, lhs, rhs)

		case OpShl:
			rhs := i.pop()
			lhs := i.pop()
			DoShiftLeft(i, lhs, rhs)

		case OpShr:
			rhs := i.pop()
			lhs := i.pop()
			DoShiftRight(i, lhs, rhs)

		case OpCmpLt:
			rhs := i.pop()
			lhs := i.pop()
			DoCmpLt(i, lhs, rhs)

		case OpCmpLte:
			rhs := i.pop()
			lhs := i.pop()
			DoCmpLte(i, lhs, rhs)

		case OpCmpGt:
			rhs := i.pop()
			lhs := i.pop()
			DoCmpGt(i, lhs, rhs)

		case OpCmpGte:
			rhs := i.pop()
			lhs := i.pop()
			DoCmpGte(i, lhs, rhs)

		case OpCmpEq:
			rhs := i.pop()
			lhs := i.pop()
			DoCmpEq(i, lhs, rhs)

		case OpCmpNe:
			rhs := i.pop()
			lhs := i.pop()
			DoCmpNe(i, lhs, rhs)

		case OpAnd:
			rhs := i.pop()
			lhs := i.pop()
			DoAnd(i, lhs, rhs)

		case OpOr:
			rhs := i.pop()
			lhs := i.pop()
			DoOr(i, lhs, rhs)

		case OpXor:
			rhs := i.pop()
			lhs := i.pop()
			DoXor(i, lhs, rhs)

		case OpStoreGlobal:
			index := ReadInt(code.Code, offsetStart)
			value := i.pop()
			code.Env0[index] = value
			forward(4)

		case OpStoreCapture:
			index := ReadInt(code.Code, offsetStart)
			value := i.pop()
			function := i.peek().Value.(*AtomCode)
			function.Env1[index] = value
			forward(4)

		case OpStoreLocal:
			index := ReadInt(code.Code, offsetStart)
			value := i.pop()
			code.Env0[index] = value
			forward(4)

		case OpJumpIfFalseOrPop:
			offset := ReadInt(code.Code, offsetStart)
			forward(4)
			value := i.peek()
			if !CoerceToBool(value) {
				jump(offset)
			} else {
				i.pop()
			}

		case OpJumpIfTrueOrPop:
			offset := ReadInt(code.Code, offsetStart)
			forward(4)
			value := i.peek()
			if CoerceToBool(value) {
				jump(offset)
			} else {
				i.pop()
			}

		case OpPopJumpIfFalse:
			offset := ReadInt(code.Code, offsetStart)
			forward(4)
			value := i.pop()
			if !CoerceToBool(value) {
				jump(offset)
			}

		case OpPopJumpIfTrue:
			offset := ReadInt(code.Code, offsetStart)
			forward(4)
			value := i.pop()
			if CoerceToBool(value) {
				jump(offset)
			}

		case OpJump:
			offset := ReadInt(code.Code, offsetStart)
			forward(4)
			jump(offset)

		case OpDupTop:
			i.pushVal(i.peek())

		case OpPopTop:
			t := i.pop()
			fmt.Println("PopTop", t.String())

		case OpReturn:
			return

		default:
			panic(fmt.Sprintf("Unknown opcode: %d", opCode))
		}
	}
}

func (i *AtomInterpreter) Interpret(atomFunc *AtomValue) {
	// Run while the frame is not empty
	i.executeFrame(nil, atomFunc, 0)

	// Dump stack
	i.EvaluationStack.Dump()
}
