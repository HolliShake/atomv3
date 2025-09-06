package runtime

import "fmt"

type AtomInterpreter struct {
	state           *AtomState
	Frame           *AtomStack
	EvaluationStack *AtomStack
}

func NewInterpreter(state *AtomState) *AtomInterpreter {
	return &AtomInterpreter{
		state:           state,
		Frame:           NewAtomStack(),
		EvaluationStack: NewAtomStack(),
	}
}

func (i *AtomInterpreter) executeFrame(frame *AtomValue, offset int) {
	// Frame here is a function

	offsetStart := offset

	code := frame.Value.(*AtomCode)
	size := len(code.OpCodes)

	forward := func(offset int) {
		offsetStart += offset
	}

	for offsetStart < size {
		opCode := code.OpCodes[offsetStart]
		offsetStart++
		switch opCode {
		case OpLoadInt:
			value := ReadInt(code.OpCodes, offsetStart)
			i.EvaluationStack.Push(
				NewAtomValueInt(value),
			)
			forward(4)

		case OpLoadNum:
			value := ReadNum(code.OpCodes, offsetStart)
			i.EvaluationStack.Push(
				NewAtomValueNum(value),
			)
			forward(8)

		case OpLoadStr:
			value := ReadStr(code.OpCodes, offsetStart)
			i.EvaluationStack.Push(
				NewAtomValueStr(value),
			)
			forward(len(value) + 1)

		case OpLoadNull:
			i.EvaluationStack.Push(
				i.state.NullValue,
			)

		case OpLoadFunction:
			offset := ReadInt(code.OpCodes, offsetStart)
			fn := i.state.FunctionTable.Get(offset)
			i.EvaluationStack.Push(fn)
			forward(4)

		case OpCall:
			argc := ReadInt(code.OpCodes, offsetStart)
			call := i.EvaluationStack.Pop()
			DoCall(i, call, argc)
			forward(4)

		case OpLoadLocal:
			index := ReadInt(code.OpCodes, offsetStart)
			value := code.Locals[index]
			i.EvaluationStack.Push(value)
			forward(4)

		case OpMul:
			rhs := i.EvaluationStack.Pop()
			lhs := i.EvaluationStack.Pop()
			DoMultiplication(i, lhs, rhs)

		case OpDiv:
			rhs := i.EvaluationStack.Pop()
			lhs := i.EvaluationStack.Pop()
			DoDivision(i, lhs, rhs)

		case OpMod:
			rhs := i.EvaluationStack.Pop()
			lhs := i.EvaluationStack.Pop()
			DoModulus(i, lhs, rhs)

		case OpAdd:
			rhs := i.EvaluationStack.Pop()
			lhs := i.EvaluationStack.Pop()
			DoAddition(i, lhs, rhs)

		case OpSub:
			rhs := i.EvaluationStack.Pop()
			lhs := i.EvaluationStack.Pop()
			DoSubtraction(i, lhs, rhs)

		case OpStoreLocal:
			index := ReadInt(code.OpCodes, offsetStart)
			value := i.EvaluationStack.Pop()
			code.Locals[index] = value
			forward(4)

		case OpPopTop:
			v := i.EvaluationStack.Pop()
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
