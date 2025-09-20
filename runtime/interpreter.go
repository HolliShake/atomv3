package runtime

import (
	"fmt"
)

const (
	TRESHOLD = 1000
)

type AtomInterpreter struct {
	State       *AtomState
	Scheduler   *AtomScheduler
	ModuleTable map[string]*AtomValue
}

func NewInterpreter(state *AtomState) *AtomInterpreter {
	interpreter := &AtomInterpreter{
		State:       state,
		ModuleTable: map[string]*AtomValue{},
	}
	interpreter.Scheduler = NewAtomScheduler(interpreter)
	return interpreter
}

func (i *AtomInterpreter) ExecuteFrame(frame *AtomCallFrame) {
	// Frame here is a function
	var code = frame.Fn.Value.(*AtomCode)
	var size = len(code.Code)
	var env0 = frame.Env
	var strt = frame.Ip

	i.Scheduler.Running(frame)

	var forwardPc = func(offset int) {
		frame.Pc += offset
	}

	var forwardIp = func(offset int) {
		strt += offset
		frame.Ip += offset
	}

	var jump = func(offset int) {
		strt = offset
		frame.Pc = offset
		frame.Ip = offset
	}

	for strt < size {
		opCode := code.Code[strt]
		forwardPc(1)
		forwardIp(1)

		switch opCode {
		case OpExportGlobal:
			DoExportGlobal(i, frame)

		case OpLoadInt:
			value := ReadInt(code.Code, strt)
			frame.Stack.Push(NewAtomValueInt(value))
			forwardIp(4)

		case OpLoadNum:
			value := ReadNum(code.Code, strt)
			frame.Stack.Push(NewAtomValueNum(value))
			forwardIp(8)

		case OpLoadStr:
			value := ReadStr(code.Code, strt)
			frame.Stack.Push(NewAtomValueStr(value))
			forwardIp(len(value) + 1)

		case OpLoadBool:
			if ReadInt(code.Code, strt) != 0 {
				frame.Stack.Push(i.State.TrueValue)
			} else {
				frame.Stack.Push(i.State.FalseValue)
			}
			forwardIp(4)

		case OpLoadNull:
			frame.Stack.Push(i.State.NullValue)

		case OpLoadArray:
			size := ReadInt(code.Code, strt)
			DoLoadArray(frame, size)
			forwardIp(4)

		case OpLoadObject:
			size := ReadInt(code.Code, strt)
			DoLoadObject(frame, size)
			forwardIp(4)

		case OpLoadName:
			name := ReadStr(code.Code, strt)
			DoLoadName(frame, env0, name)
			forwardIp(len(name) + 1)

		case OpLoadModule0:
			name := ReadStr(code.Code, strt)
			DoLoadModule0(i, frame, name)
			forwardIp(len(name) + 1)

		case OpLoadModule1:
			path := ReadStr(code.Code, strt)
			DoLoadModule1(i, frame, path)
			forwardIp(len(path) + 1)

		case OpLoadFunction:
			offset := ReadInt(code.Code, strt)
			DoLoadFunction(i, frame, offset)
			forwardIp(4)

		case OpMakeClass:
			size := ReadInt(code.Code, strt)
			name := ReadStr(code.Code, strt+4)
			DoMakeClass(i, frame, name, size)
			forwardIp(4 + len(name) + 1)

		case OpExtendClass:
			ext := frame.Stack.Pop()
			cls := frame.Stack.Peek()
			DoExtendClass(cls, ext)

		case OpMakeEnum:
			size := ReadInt(code.Code, strt)
			DoMakeEnum(frame, env0, size)
			forwardIp(4)

		case OpCallConstructor:
			argc := ReadInt(code.Code, strt)
			call := frame.Stack.Pop()
			DoCallConstructor(i, frame, env0, call, argc)
			forwardIp(4)

		case OpCall:
			argc := ReadInt(code.Code, strt)
			call := frame.Stack.Pop()
			DoCall(i, frame, env0, call, argc)
			forwardIp(4)

		case OpAwait:
			if !CheckType(frame.Stack.Peek(), AtomTypePromise) {
				continue
			}
			if i.Scheduler.Await(frame) {
				return
			}

		case OpNot:
			val := frame.Stack.Pop()
			DoNot(i, frame, val)

		case OpNeg:
			val := frame.Stack.Pop()
			DoNeg(frame, val)

		case OpPos:
			val := frame.Stack.Pop()
			DoPos(frame, val)

		case OpTypeof:
			val := frame.Stack.Pop()
			DoTypeof(frame, val)

		case OpIndex:
			idx := frame.Stack.Pop()
			obj := frame.Stack.Pop()
			DoIndex(i, frame, obj, idx)

		case OpPluckAttribute:
			att := ReadStr(code.Code, strt)
			obj := frame.Stack.Peek()
			DoPluckAttribute(i, frame, obj, att)
			forwardIp(len(att) + 1)

		case OpMul:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoMultiplication(frame, lhs, rhs)

		case OpDiv:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoDivision(frame, lhs, rhs)

		case OpMod:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoModulus(frame, lhs, rhs)

		case OpAdd:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoAddition(frame, lhs, rhs)

		case OpSub:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoSubtraction(frame, lhs, rhs)

		case OpShl:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoShiftLeft(frame, lhs, rhs)

		case OpShr:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoShiftRight(frame, lhs, rhs)

		case OpCmpLt:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoCmpLt(i, frame, lhs, rhs)

		case OpCmpLte:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoCmpLte(i, frame, lhs, rhs)

		case OpCmpGt:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoCmpGt(i, frame, lhs, rhs)

		case OpCmpGte:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoCmpGte(i, frame, lhs, rhs)

		case OpCmpEq:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoCmpEq(i, frame, lhs, rhs)

		case OpCmpNe:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoCmpNe(i, frame, lhs, rhs)

		case OpAnd:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoAnd(frame, lhs, rhs)

		case OpOr:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoOr(frame, lhs, rhs)

		case OpXor:
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Pop()
			DoXor(frame, lhs, rhs)

		case OpStoreModule:
			name := ReadStr(code.Code, strt)
			DoStoreModule(i, frame, name)
			forwardIp(len(name) + 1)

		case OpInitVar:
			v := ReadStr(code.Code, strt)
			g := code.Code[strt+len(v)+1] == 1
			c := code.Code[strt+len(v)+2] == 1
			value := frame.Stack.Pop()
			err := env0.New(v, g, c, value)
			if err != nil {
				panic(err)
			}
			forwardIp(len(v) + 1 + 1 + 1)

		case OpStoreFast:
			param := ReadStr(code.Code, strt)
			value := frame.Stack.Pop()
			env0.New(param, false, false, value)
			forwardIp(len(param) + 1)

		case OpStoreLocal:
			index := ReadStr(code.Code, strt)
			value := frame.Stack.Pop()
			env0.Store(index, value)
			forwardIp(len(index) + 1)

		case OpSetIndex:
			idx := frame.Stack.Pop()
			obj := frame.Stack.Pop()
			DoSetIndex(i, frame, obj, idx)

		case OpJumpIfFalseOrPop:
			offset := ReadInt(code.Code, strt)
			forwardIp(4)
			value := frame.Stack.Peek()
			if !CoerceToBool(value) {
				jump(offset)
			} else {
				frame.Stack.Pop()
			}

		case OpJumpIfTrueOrPop:
			offset := ReadInt(code.Code, strt)
			forwardIp(4)
			value := frame.Stack.Peek()
			if CoerceToBool(value) {
				jump(offset)
			} else {
				frame.Stack.Pop()
			}

		case OpPopJumpIfFalse:
			offset := ReadInt(code.Code, strt)
			forwardIp(4)
			value := frame.Stack.Pop()
			if !CoerceToBool(value) {
				jump(offset)
			}

		case OpPopJumpIfTrue:
			offset := ReadInt(code.Code, strt)
			forwardIp(4)
			value := frame.Stack.Pop()
			if CoerceToBool(value) {
				jump(offset)
			}

		case OpPeekJumpIfEqual:
			offset := ReadInt(code.Code, strt)
			forwardIp(4)
			rhs := frame.Stack.Pop()
			lhs := frame.Stack.Peek()
			if rhs.HashValue() == lhs.HashValue() {
				jump(offset)
			}

		case OpPopJumpIfNotError:
			offset := ReadInt(code.Code, strt)
			forwardIp(4)
			value := frame.Stack.Peek()
			if !CheckType(value, AtomTypeErr) {
				jump(offset)
			}

		case OpJump:
			offset := ReadInt(code.Code, strt)
			forwardIp(4)
			jump(offset)

		case OpAbsoluteJump:
			// Alias for OpJump
			offset := ReadInt(code.Code, strt)
			forwardIp(4)
			jump(offset)

		case OpEnterBlock:
			env0 = NewAtomEnv(env0)

		case OpExitBlock:
			env0 = env0.Parent

		case OpDupTop:
			frame.Stack.Push(frame.Stack.Peek())

		case OpNoOp:
			forwardIp(0)

		case OpPopTop:
			frame.Stack.Pop()

		case OpReturn:
			i.Scheduler.Resolve(frame)
			return

		default:
			// fmt.Println(Decompile(code))
			panic(fmt.Sprintf("%s:: Unknown opcode: %d at %d", frame.Fn.Value.(*AtomCode).Name, opCode, strt))
		}
	}
}

func (i *AtomInterpreter) Interpret(atomFunc *AtomValue) {
	DefineModule(i, "std", EXPORT_STD)
	DefineModule(i, "object", EXPORT_OBJECT)

	// Run while the frame is not empty
	i.ExecuteFrame(NewAtomCallFrame(nil, atomFunc, NewAtomEnv(nil), 0))

	i.Scheduler.Run()
}
