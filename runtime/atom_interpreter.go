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

		case OpMul:
			lhs := i.EvaluationStack.Pop()
			rhs := i.EvaluationStack.Pop()
			DoMultiplication(i, lhs, rhs)

		case OpDiv:
			lhs := i.EvaluationStack.Pop()
			rhs := i.EvaluationStack.Pop()
			DoDivision(i, lhs, rhs)

		case OpMod:
			lhs := i.EvaluationStack.Pop()
			rhs := i.EvaluationStack.Pop()
			DoModulus(i, lhs, rhs)

		case OpAdd:
			lhs := i.EvaluationStack.Pop()
			rhs := i.EvaluationStack.Pop()
			DoAddition(i, lhs, rhs)

		case OpSub:
			lhs := i.EvaluationStack.Pop()
			rhs := i.EvaluationStack.Pop()
			DoSubtraction(i, lhs, rhs)

		case OpReturn:
			i.Frame.Pop()
			return

		default:
			panic(fmt.Sprintf("Unknown opcode: %d", opCode))
		}
	}
}

func (i *AtomInterpreter) Interpret(atomFunc *AtomValue) {
	i.Frame.Push(atomFunc)

	// Run while the frame is not empty
	for i.Frame.Len() > 0 {
		i.executeFrame(i.Frame.Peek(), 0)
	}

	// While has pending frames for async

	// Dump stack
	i.EvaluationStack.Dump()
}
