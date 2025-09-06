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

func (i *AtomInterpreter) pushValue(value *AtomValue) {
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

func (i *AtomInterpreter) executeFrame(frame *AtomValue, offset int) {
	// Frame here is a function

	offsetStart := offset

	code := frame.Value.(*AtomCode)
	size := len(code.OpCodes)

	forward := func(offset int) {
		offsetStart += offset
	}

	jump := func(offset int) {
		offsetStart = offset
	}

	for offsetStart < size {
		opCode := code.OpCodes[offsetStart]
		offsetStart++

		if (i.Allocation % TRESHOLD) == 0 {
			// TODO: Garbage collection
		}

		switch opCode {
		case OpLoadInt:
			value := ReadInt(code.OpCodes, offsetStart)
			i.pushValue(
				NewAtomValueInt(value),
			)
			forward(4)

		case OpLoadNum:
			value := ReadNum(code.OpCodes, offsetStart)
			i.pushValue(
				NewAtomValueNum(value),
			)
			forward(8)

		case OpLoadStr:
			value := ReadStr(code.OpCodes, offsetStart)
			i.pushValue(
				NewAtomValueStr(value),
			)
			forward(len(value) + 1)

		case OpLoadNull:
			i.pushValue(
				i.state.NullValue,
			)

		case OpLoadFunction:
			offset := ReadInt(code.OpCodes, offsetStart)
			fn := i.state.FunctionTable.Get(offset)
			i.pushRef(fn)
			forward(4)

		case OpCall:
			argc := ReadInt(code.OpCodes, offsetStart)
			forward(4)
			call := i.pop()
			DoCall(i, call, argc)

		case OpLoadLocal:
			index := ReadInt(code.OpCodes, offsetStart)
			value := code.Locals[index]
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

		case OpStoreLocal:
			index := ReadInt(code.OpCodes, offsetStart)
			value := i.pop()
			code.Locals[index] = value
			forward(4)

		case OpJumpIfTrueOrPop:
			offset := ReadInt(code.OpCodes, offsetStart)
			forward(4)
			value := i.peek()
			if CoerceToBool(value) {
				jump(offset)
			} else {
				i.pop()
			}

		case OpPopJumpIfFalse:
			offset := ReadInt(code.OpCodes, offsetStart)
			forward(4)
			value := i.pop()
			if !CoerceToBool(value) {
				jump(offset)
			}

		case OpJump:
			offset := ReadInt(code.OpCodes, offsetStart)
			jump(offset)

		case OpPopTop:
			v := i.pop()
			fmt.Println("PopTop", v.String())

		case OpReturn:
			return

		default:
			panic(fmt.Sprintf("Unknown opcode: %d", opCode))
		}
	}
}

func (i *AtomInterpreter) Interpret(atomFunc *AtomValue) {
	i.Frame.Push(atomFunc)

	// Run while the frame is not empty
	i.executeFrame(atomFunc, 0)

	// While has pending frames for async

	// Dump stack
	i.EvaluationStack.Dump()
}
