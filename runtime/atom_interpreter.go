package runtime

import (
	"fmt"
)

const (
	TRESHOLD = 1000
)

type AtomInterpreter struct {
	State           *AtomState
	Frame           *AtomStack
	EvaluationStack *AtomStack
	ModuleTable     map[string]*AtomValue
	GcRoot          *AtomValue
	Allocation      int
}

func NewInterpreter(state *AtomState) *AtomInterpreter {
	return &AtomInterpreter{
		State:           state,
		Frame:           NewAtomStack(),
		EvaluationStack: NewAtomStack(),
		GcRoot:          NewAtomValue(AtomTypeObj),
		ModuleTable:     map[string]*AtomValue{},
	}
}

func (i *AtomInterpreter) pushVal(value *AtomValue) {
	// i.GcRoot.Next = value
	// i.GcRoot = value
	i.EvaluationStack.Push(value)
	// i.Allocation++
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

		case OpLoadBool:
			value := ReadInt(code.Code, offsetStart)
			if value != 0 {
				i.pushRef(i.State.TrueValue)
			} else {
				i.pushRef(i.State.FalseValue)
			}
			forward(4)

		case OpLoadNull:
			i.pushRef(
				i.State.NullValue,
			)

		case OpLoadArray:
			length := ReadInt(code.Code, offsetStart)
			elements := []*AtomValue{}
			for range length {
				elements = append(elements, i.pop())
			}
			i.pushVal(
				NewAtomValueArray(elements),
			)
			forward(4)

		case OpLoadObject:
			length := ReadInt(code.Code, offsetStart)
			elements := map[string]*AtomValue{}
			for range length {
				k := i.pop()
				v := i.pop()
				elements[k.Value.(string)] = v
			}
			i.pushVal(
				NewAtomValueObject(elements),
			)
			forward(4)

		case OpLoadFunction:
			offset := ReadInt(code.Code, offsetStart)
			fn := i.State.FunctionTable.Get(offset)
			i.pushRef(fn)
			forward(4)

		case OpLoadModule0:
			name := ReadStr(code.Code, offsetStart)
			DoLoadModule0(i, name)
			forward(len(name) + 1)

		case OpIndex:
			index := i.pop()
			obj := i.pop()
			DoIndex(i, obj, index)

		case OpPluckAttribute:
			attribute := ReadStr(code.Code, offsetStart)
			obj := i.peek()
			DoPluckAttribute(i, obj, attribute)
			forward(len(attribute) + 1)

		case OpCall:
			argc := ReadInt(code.Code, offsetStart)
			forward(4)
			call := i.pop()
			DoCall(i, call, argc)

		case OpLoadLocal:
			index := ReadInt(code.Code, offsetStart)
			value := code.Env0[index]
			i.pushRef(value.Get())
			forward(4)

		case OpNot:
			val := i.pop()
			DoNot(i, val)

		case OpPos:
			val := i.pop()
			DoPos(i, val)

		case OpNeg:
			val := i.pop()
			DoNeg(i, val)

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
			// Alias for OpStoreLocal
			index := ReadInt(code.Code, offsetStart)
			value := i.pop()
			code.Env0[index].Set(value)
			forward(4)

		case OpStoreLocal:
			index := ReadInt(code.Code, offsetStart)
			value := i.pop()
			code.Env0[index].Set(value)
			forward(4)

		case OpSetIndex:
			index := i.pop()
			obj := i.pop()
			DoSetIndex(i, obj, index)

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

		case OpPeekJumpIfEqual:
			offset := ReadInt(code.Code, offsetStart)
			forward(4)
			rhs := i.pop()
			lhs := i.peek()
			if rhs.Value == lhs.Value {
				jump(offset)
			}

		case OpJump:
			offset := ReadInt(code.Code, offsetStart)
			forward(4)
			jump(offset)

		case OpAbsoluteJump:
			// Alias for OpJump
			offset := ReadInt(code.Code, offsetStart)
			forward(4)
			jump(offset)

		case OpPopJumpIfNotError:
			offset := ReadInt(code.Code, offsetStart)
			forward(4)
			value := i.peek()
			if !CheckType(value, AtomTypeErr) {
				jump(offset)
			}

		case OpDupTop:
			i.pushVal(i.peek())

		case OpNoOp:
			forward(0)

		case OpPopTop:
			i.pop()

		case OpReturn:
			return

		default:
			panic(fmt.Sprintf("Unknown opcode: %d", opCode))
		}
	}
}

func (i *AtomInterpreter) Interpret(atomFunc *AtomValue) {
	DefineModule(i, "std", EXPORT_STD)

	// Run while the frame is not empty
	i.executeFrame(atomFunc, 0)

	if i.EvaluationStack.Len() != 1 {
		i.EvaluationStack.Dump()
		panic(fmt.Sprintf("Evaluation stack is not empty: %d", i.EvaluationStack.Len()))
	}
}
