package runtime

import (
	"fmt"
	"strings"
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

func StackTrace(frame *AtomCallFrame) string {
	builder := strings.Builder{}
	builder.WriteByte('\n')

	current := frame
	for current != nil {
		builder.WriteString(current.Fn.Obj.(*AtomCode).Name)
		if current.Caller != nil {
			builder.WriteString("\n")
		}
		current = current.Caller
	}

	return builder.String()
}

func (i *AtomInterpreter) ExecuteFrame(frame *AtomCallFrame) {
	// Frame here is a function
	var code = frame.Fn.Obj.(*AtomCode)
	var size = len(code.Code)
	var strt = frame.Ip

	i.Scheduler.Running(frame)

	var forwardIp = func(offset int) {
		strt += offset
		frame.Ip += offset
	}

	var jump = func(offset int) {
		strt = offset
		frame.Ip = offset
	}

	var stringCache = map[int]string{}
	var intCache = map[int]int{}
	var numCache = map[int]float64{}

	var writeString = func(offset int) string {
		save := stringCache[offset]
		if save == "" {
			save = ReadStr(code.Code, offset)
			stringCache[offset] = save
		}
		return save
	}

	var writeInt = func(offset int) int {
		save := intCache[offset]
		if save == 0 {
			save = ReadInt(code.Code, offset)
			intCache[offset] = save
		}
		return save
	}

	var writeNum = func(offset int) float64 {
		save := numCache[offset]
		if save == 0 {
			save = ReadNum(code.Code, offset)
			numCache[offset] = save
		}
		return save
	}

	for strt < size {
		opCode := code.Code[strt]
		forwardIp(1)

		switch opCode {
		case OpMakeModule:
			size := ReadInt(code.Code, strt)
			DoMakeModule(i, frame, size)
			forwardIp(4)

		case OpLoadInt:
			value := writeInt(strt)
			frame.Stack.Push(NewAtomValueInt(value))
			forwardIp(4)

		case OpLoadBigInt:
			value := writeString(strt)
			frame.Stack.Push(NewAtomValueBigInt(BigInt(value)))
			forwardIp(len(value) + 1)

		case OpLoadNum:
			value := writeNum(strt)
			frame.Stack.Push(NewAtomValueNum(value))
			forwardIp(8)

		case OpLoadStr:
			value := writeString(strt)
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

		case OpLoadBase:
			DoLoadBase(i, frame)

		case OpLoadArray:
			size := ReadInt(code.Code, strt)
			DoLoadArray(frame, size)
			forwardIp(4)

		case OpLoadObject:
			size := ReadInt(code.Code, strt)
			DoLoadObject(frame, size)
			forwardIp(4)

		case OpLoadName:
			index := writeString(strt)
			DoLoadName(frame, index)
			forwardIp(len(index) + 1)

		case OpLoadModule:
			name := writeString(strt)
			DoLoadModule(i, frame, name)
			forwardIp(len(name) + 1)

		case OpLoadFunction:
			offset := ReadInt(code.Code, strt)
			DoLoadFunction(i, frame, offset)
			forwardIp(4)

		case OpMakeClass:
			size := ReadInt(code.Code, strt)
			name := writeString(strt + 4)
			DoMakeClass(i, frame, name, size)
			forwardIp(4 + len(name) + 1)

		case OpExtendClass:
			ext := frame.Stack.Pop()
			cls := frame.Stack.Peek()
			DoExtendClass(cls, ext)

		case OpMakeEnum:
			size := ReadInt(code.Code, strt)
			DoMakeEnum(frame, size)
			forwardIp(4)

		case OpCallConstructor:
			argc := ReadInt(code.Code, strt)
			call := frame.Stack.Pop()
			DoCallConstructor(i, frame, call, argc)
			forwardIp(4)

		case OpCall:
			argc := ReadInt(code.Code, strt)
			call := frame.Stack.Pop()
			DoCall(i, frame, call, argc)
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

		case OpInc:
			val := frame.Stack.Pop()
			DoInc(frame, val)

		case OpDec:
			val := frame.Stack.Pop()
			DoDec(frame, val)

		case OpTypeof:
			val := frame.Stack.Pop()
			DoTypeof(frame, val)

		case OpIndex:
			idx := frame.Stack.Pop()
			obj := frame.Stack.Pop()
			DoIndex(i, frame, obj, idx)

		case OpPluckAttribute:
			att := writeString(strt)
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
			name := writeString(strt)
			DoStoreModule(i, frame, name)
			forwardIp(len(name) + 1)

		case OpInitLocal:
			index := writeString(strt)
			value := frame.Stack.Pop()
			DoInitLocal(i, frame, index, value)
			forwardIp(len(index) + 1)

		case OpStoreLocal:
			index := writeString(strt)
			value := frame.Stack.Pop()
			DoStoreLocal(i, frame, index, value)
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

		case OpDupTop:
			frame.Stack.Push(frame.Stack.Peek())

		case OpDupTop2:
			a := frame.Stack.Pop()
			b := frame.Stack.Pop()
			frame.Stack.Push(b)
			frame.Stack.Push(a)
			frame.Stack.Push(b)
			frame.Stack.Push(a)

		case OpNoOp:
			forwardIp(0)

		case OpPopTop:
			frame.Stack.Pop()

		case OpEnterBlock:
			_ = ReadInt(code.Code, strt)
			forwardIp(4)
			frame.Env = NewAtomEnv(frame.Env)

		case OpExitBlock:
			depth := ReadInt(code.Code, strt)
			forwardIp(4)
			for range depth {
				frame.Env = frame.Env.Parent
			}

		case OpRot2:
			DoRot2(frame)

		case OpRot3:
			DoRot3(frame)

		case OpRot4:
			DoRot4(frame)

		case OpReturn:
			if frame.Stack.Len() != 1 {
				frame.Stack.Dump()
				panic(fmt.Sprintf("%s: Return with more than 1 value on the stack %d", frame.Fn.Obj.(*AtomCode).Name, frame.Stack.Len()))
			}
			i.Scheduler.Resolve(frame)
			return

		default:
			// fmt.Println(Decompile(code))
			panic(fmt.Sprintf("%s:: Unknown opcode: %d at %d", frame.Fn.Obj.(*AtomCode).Name, opCode, strt))
		}
	}
}

func (i *AtomInterpreter) Interpret(atomFunc *AtomValue) {
	DefineModule(i, "std", EXPORT_STD)
	DefineModule(i, "object", EXPORT_OBJECT)
	DefineModule(i, "math", EXPORT_MATH)
	DefineModule(i, "path", EXPORT_PATH)
	DefineModule(i, "os", EXPORT_OS)
	DefineModule(i, "file", EXPORT_FILE)
	DefineModule(i, "string", EXPORT_STRING)
	DefineModule(i, "number", EXPORT_NUMBER)

	// Run while the frame is not empty
	i.ExecuteFrame(NewAtomCallFrame(nil, atomFunc, 0))

	i.Scheduler.Run()
}
