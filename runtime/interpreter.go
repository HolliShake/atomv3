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
	MicroTask       *Queue[*AtomCallFrame]
	ModuleTable     map[string]*AtomValue
}

func NewInterpreter(state *AtomState) *AtomInterpreter {
	return &AtomInterpreter{
		State:           state,
		Frame:           NewAtomStack(),
		EvaluationStack: NewAtomStack(),
		MicroTask:       NewQueue[*AtomCallFrame](),
		ModuleTable:     map[string]*AtomValue{},
	}
}

func (i *AtomInterpreter) pushVal(value *AtomValue) {
	i.EvaluationStack.Push(value)
}

func (i *AtomInterpreter) popp() *AtomValue {
	return i.EvaluationStack.Pop()
}

func (i *AtomInterpreter) peek() *AtomValue {
	return i.EvaluationStack.Peek()
}

func (i *AtomInterpreter) executeFrame(callFrame *AtomCallFrame) {
	// Frame here is a function
	var strt = callFrame.Ip
	var code = callFrame.Fn.Value.(*AtomCode)
	var size = len(code.Code)
	var env0 = callFrame.Env

	var forward = func(offset int) {
		strt += offset
	}

	var jump = func(offset int) {
		strt = offset
	}

	var saveAsTask = func(promise *AtomValue, stack int) {
		callFrame.Ip = strt
		callFrame.Stack = stack
		callFrame.Promise = promise
		i.MicroTask.Enqueue(callFrame)
	}

	for strt < size {
		opCode := code.Code[strt]
		strt++

		switch opCode {
		case OpLoadInt:
			value := ReadInt(code.Code, strt)
			i.pushVal(NewAtomValueInt(value))
			forward(4)

		case OpLoadNum:
			value := ReadNum(code.Code, strt)
			i.pushVal(NewAtomValueNum(value))
			forward(8)

		case OpLoadStr:
			value := ReadStr(code.Code, strt)
			i.pushVal(NewAtomValueStr(value))
			forward(len(value) + 1)

		case OpLoadBool:
			if ReadInt(code.Code, strt) != 0 {
				i.pushVal(i.State.TrueValue)
			} else {
				i.pushVal(i.State.FalseValue)
			}
			forward(4)

		case OpLoadNull:
			i.pushVal(i.State.NullValue)

		case OpLoadArray:
			size := ReadInt(code.Code, strt)
			DoLoadArray(i, size)
			forward(4)

		case OpLoadObject:
			size := ReadInt(code.Code, strt)
			DoLoadObject(i, size)
			forward(4)

		case OpLoadName:
			variable := ReadStr(code.Code, strt)
			DoLoadName(i, env0, variable)
			forward(len(variable) + 1)

		case OpLoadModule0:
			name := ReadStr(code.Code, strt)
			DoLoadModule0(i, name)
			forward(len(name) + 1)

		case OpLoadFunction:
			offset := ReadInt(code.Code, strt)
			DoLoadFunction(i, offset)
			forward(4)

		case OpMakeClass:
			size := ReadInt(code.Code, strt)
			name := ReadStr(code.Code, strt+4)
			DoMakeClass(i, name, size)
			forward(4 + len(name) + 1)

		case OpExtendClass:
			ext := i.popp()
			cls := i.peek()
			DoExtendClass(i, cls, ext)

		case OpMakeEnum:
			size := ReadInt(code.Code, strt)
			DoMakeEnum(i, env0, size)
			forward(4)

		case OpCallConstructor:
			argc := ReadInt(code.Code, strt)
			call := i.popp()
			DoCallConstructor(i, env0, call, argc)
			forward(4)

		case OpCall:
			argc := ReadInt(code.Code, strt)
			call := i.popp()
			DoCall(i, env0, call, argc)
			forward(4)

		case OpAwaitCall:
			argc := ReadInt(code.Code, strt)
			call := i.popp()
			DoCall(i, env0, call, argc)

			if !CheckType(i.peek(), AtomTypePromise) {
				forward(4)
				continue
			}

			// Is fullfilled?
			if i.peek().Value.(*AtomPromise).IsFulfilled() {
				// unwrap
				i.pushVal(i.popp().Value.(*AtomPromise).Value)
			}

			stack := i.EvaluationStack.Index

			// Push a promise
			prom := NewAtomValuePromise(PromiseStatePending, nil)
			i.pushVal(prom)

			forward(4)
			saveAsTask(prom, stack)
			// Do immediate return
			return

		case OpNot:
			val := i.popp()
			DoNot(i, val)

		case OpNeg:
			val := i.popp()
			DoNeg(i, val)

		case OpPos:
			val := i.popp()
			DoPos(i, val)

		case OpTypeof:
			val := i.popp()
			DoTypeof(i, val)

		case OpIndex:
			index := i.popp()
			obj := i.popp()
			DoIndex(i, obj, index)

		case OpPluckAttribute:
			attribute := ReadStr(code.Code, strt)
			obj := i.peek()
			DoPluckAttribute(i, obj, attribute)
			forward(len(attribute) + 1)

		case OpMul:
			rhs := i.popp()
			lhs := i.popp()
			DoMultiplication(i, lhs, rhs)

		case OpDiv:
			rhs := i.popp()
			lhs := i.popp()
			DoDivision(i, lhs, rhs)

		case OpMod:
			rhs := i.popp()
			lhs := i.popp()
			DoModulus(i, lhs, rhs)

		case OpAdd:
			rhs := i.popp()
			lhs := i.popp()
			DoAddition(i, lhs, rhs)

		case OpSub:
			rhs := i.popp()
			lhs := i.popp()
			DoSubtraction(i, lhs, rhs)

		case OpShl:
			rhs := i.popp()
			lhs := i.popp()
			DoShiftLeft(i, lhs, rhs)

		case OpShr:
			rhs := i.popp()
			lhs := i.popp()
			DoShiftRight(i, lhs, rhs)

		case OpCmpLt:
			rhs := i.popp()
			lhs := i.popp()
			DoCmpLt(i, lhs, rhs)

		case OpCmpLte:
			rhs := i.popp()
			lhs := i.popp()
			DoCmpLte(i, lhs, rhs)

		case OpCmpGt:
			rhs := i.popp()
			lhs := i.popp()
			DoCmpGt(i, lhs, rhs)

		case OpCmpGte:
			rhs := i.popp()
			lhs := i.popp()
			DoCmpGte(i, lhs, rhs)

		case OpCmpEq:
			rhs := i.popp()
			lhs := i.popp()
			DoCmpEq(i, lhs, rhs)

		case OpCmpNe:
			rhs := i.popp()
			lhs := i.popp()
			DoCmpNe(i, lhs, rhs)

		case OpAnd:
			rhs := i.popp()
			lhs := i.popp()
			DoAnd(i, lhs, rhs)

		case OpOr:
			rhs := i.popp()
			lhs := i.popp()
			DoOr(i, lhs, rhs)

		case OpXor:
			rhs := i.popp()
			lhs := i.popp()
			DoXor(i, lhs, rhs)

		case OpInitVar:
			v := ReadStr(code.Code, strt)
			g := code.Code[strt+len(v)+1] == 1
			c := code.Code[strt+len(v)+2] == 1
			value := i.popp()
			err := env0.New(v, g, c, value)
			if err != nil {
				panic(err)
			}
			forward(len(v) + 1 + 1 + 1)

		case OpStoreFast:
			param := ReadStr(code.Code, strt)
			value := i.popp()
			env0.New(param, false, false, value)
			forward(len(param) + 1)

		case OpStoreLocal:
			index := ReadStr(code.Code, strt)
			value := i.popp()
			env0.Store(index, value)
			forward(len(index) + 1)

		case OpSetIndex:
			index := i.popp()
			obj := i.popp()
			DoSetIndex(i, obj, index)

		case OpJumpIfFalseOrPop:
			offset := ReadInt(code.Code, strt)
			forward(4)
			value := i.peek()
			if !CoerceToBool(value) {
				jump(offset)
			} else {
				i.popp()
			}

		case OpJumpIfTrueOrPop:
			offset := ReadInt(code.Code, strt)
			forward(4)
			value := i.peek()
			if CoerceToBool(value) {
				jump(offset)
			} else {
				i.popp()
			}

		case OpPopJumpIfFalse:
			offset := ReadInt(code.Code, strt)
			forward(4)
			value := i.popp()
			if !CoerceToBool(value) {
				jump(offset)
			}

		case OpPopJumpIfTrue:
			offset := ReadInt(code.Code, strt)
			forward(4)
			value := i.popp()
			if CoerceToBool(value) {
				jump(offset)
			}

		case OpPeekJumpIfEqual:
			offset := ReadInt(code.Code, strt)
			forward(4)
			rhs := i.popp()
			lhs := i.peek()
			if rhs.HashValue() == lhs.HashValue() {
				jump(offset)
			}

		case OpPopJumpIfNotError:
			offset := ReadInt(code.Code, strt)
			forward(4)
			value := i.peek()
			if !CheckType(value, AtomTypeErr) {
				jump(offset)
			}

		case OpJump:
			offset := ReadInt(code.Code, strt)
			forward(4)
			jump(offset)

		case OpAbsoluteJump:
			// Alias for OpJump
			offset := ReadInt(code.Code, strt)
			forward(4)
			jump(offset)

		case OpEnterBlock:
			env0 = NewAtomEnv(env0)

		case OpExitBlock:
			env0 = env0.Parent

		case OpDupTop:
			i.pushVal(i.peek())

		case OpNoOp:
			forward(0)

		case OpPopTop:
			i.popp()

		case OpReturn:
			if callFrame.Fn.Value.(*AtomCode).Async {
				// Wrap
				i.pushVal(
					NewAtomValuePromise(PromiseStateFulfilled, i.popp()),
				)
			}
			return

		default:
			panic(fmt.Sprintf("Unknown opcode: %d", opCode))
		}
	}
}

func (i *AtomInterpreter) Interpret(atomFunc *AtomValue) {
	DefineModule(i, "std", EXPORT_STD)

	// Run while the frame is not empty
	i.executeFrame(NewAtomCallFrame(
		atomFunc,
		NewAtomEnv(nil),
		0,
	))

	for !i.MicroTask.IsEmpty() {
		top, _ := i.MicroTask.Dequeue()
		i.EvaluationStack.SetIndex(top.Stack)
		i.executeFrame(top)
	}

	if i.EvaluationStack.Len() != 1 {
		i.EvaluationStack.Dump()
		panic(fmt.Sprintf("Evaluation stack is not empty: %d", i.EvaluationStack.Len()))
	}
}
