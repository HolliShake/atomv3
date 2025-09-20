package runtime

import (
	"fmt"
	"strings"
)

func Decompile(code *AtomCode) string {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("Function: %s\n", code.Name))
	builder.WriteString(fmt.Sprintf("Code Length: %d\n", len(code.Code)))
	builder.WriteString("Instructions:\n")

	pc := 0
	for pc < len(code.Code) {
		opcode := code.Code[pc]
		pc++
		builder.WriteString(fmt.Sprintf("%08d: ", pc-1))

		switch opcode {
		case OpMakeModule:
			size := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("MAKE_MODULE %d\n", size))
			pc += 4

		case OpLoadInt:
			value := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("LOAD_INT %d\n", value))
			pc += 4

		case OpLoadNum:
			value := ReadNum(code.Code, pc)
			builder.WriteString(fmt.Sprintf("LOAD_NUM %f\n", value))
			pc += 8

		case OpLoadStr:
			value := ReadStr(code.Code, pc)
			builder.WriteString(fmt.Sprintf("LOAD_STR \"%s\"\n", value))
			pc += len(value) + 1

		case OpLoadBool:
			value := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("LOAD_BOOL %t\n", value != 0))
			pc += 4

		case OpLoadNull:
			builder.WriteString("LOAD_NULL\n")

		case OpLoadArray:
			size := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("LOAD_ARRAY %d\n", size))
			pc += 4

		case OpLoadObject:
			size := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("LOAD_OBJECT %d\n", size))
			pc += 4

		case OpLoadName:
			index := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("LOAD_NAME %d\n", index))
			pc += 4

		case OpLoadCapture:
			index := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("LOAD_CAPTURE %d\n", index))
			pc += 4

		case OpLoadModule:
			name := ReadStr(code.Code, pc)
			builder.WriteString(fmt.Sprintf("LOAD_MODULE \"%s\"\n", name))
			pc += len(name) + 1

		case OpLoadFunction:
			offset := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("LOAD_FUNCTION %d\n", offset))
			pc += 4

		case OpMakeClass:
			size := ReadInt(code.Code, pc)
			name := ReadStr(code.Code, pc+4)
			builder.WriteString(fmt.Sprintf("MAKE_CLASS %d \"%s\"\n", size, name))
			pc += 4 + len(name) + 1

		case OpExtendClass:
			builder.WriteString("EXTEND_CLASS\n")

		case OpMakeEnum:
			size := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("MAKE_ENUM %d\n", size))
			pc += 4

		case OpCallConstructor:
			argc := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("CALL_CONSTRUCTOR %d\n", argc))
			pc += 4

		case OpCall:
			argc := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("CALL %d\n", argc))
			pc += 4

		case OpAwait:
			builder.WriteString("AWAIT\n")

		case OpNot:
			builder.WriteString("NOT\n")

		case OpNeg:
			builder.WriteString("NEG\n")

		case OpPos:
			builder.WriteString("POS\n")

		case OpTypeof:
			builder.WriteString("TYPEOF\n")

		case OpIndex:
			builder.WriteString("INDEX\n")

		case OpPluckAttribute:
			attr := ReadStr(code.Code, pc)
			builder.WriteString(fmt.Sprintf("PLUCK_ATTRIBUTE \"%s\"\n", attr))
			pc += len(attr) + 1

		case OpMul:
			builder.WriteString("MUL\n")

		case OpDiv:
			builder.WriteString("DIV\n")

		case OpMod:
			builder.WriteString("MOD\n")

		case OpAdd:
			builder.WriteString("ADD\n")

		case OpSub:
			builder.WriteString("SUB\n")

		case OpShl:
			builder.WriteString("SHL\n")

		case OpShr:
			builder.WriteString("SHR\n")

		case OpCmpLt:
			builder.WriteString("CMP_LT\n")

		case OpCmpLte:
			builder.WriteString("CMP_LTE\n")

		case OpCmpGt:
			builder.WriteString("CMP_GT\n")

		case OpCmpGte:
			builder.WriteString("CMP_GTE\n")

		case OpCmpEq:
			builder.WriteString("CMP_EQ\n")

		case OpCmpNe:
			builder.WriteString("CMP_NE\n")

		case OpAnd:
			builder.WriteString("AND\n")

		case OpOr:
			builder.WriteString("OR\n")

		case OpXor:
			builder.WriteString("XOR\n")

		case OpStoreModule:
			name := ReadStr(code.Code, pc)
			builder.WriteString(fmt.Sprintf("STORE_MODULE \"%s\"\n", name))
			pc += len(name) + 1

		case OpStoreCapture:
			index := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("STORE_CAPTURE %d\n", index))
			pc += 4

		case OpStoreLocal:
			index := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("STORE_LOCAL %d\n", index))
			pc += 4

		case OpSetIndex:
			builder.WriteString("SET_INDEX\n")

		case OpJumpIfFalseOrPop:
			offset := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("JUMP_IF_FALSE_OR_POP %d\n", offset))
			pc += 4

		case OpJumpIfTrueOrPop:
			offset := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("JUMP_IF_TRUE_OR_POP %d\n", offset))
			pc += 4

		case OpPopJumpIfFalse:
			offset := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("POP_JUMP_IF_FALSE %d\n", offset))
			pc += 4

		case OpPopJumpIfTrue:
			offset := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("POP_JUMP_IF_TRUE %d\n", offset))
			pc += 4

		case OpPeekJumpIfEqual:
			offset := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("PEEK_JUMP_IF_EQUAL %d\n", offset))
			pc += 4

		case OpPopJumpIfNotError:
			offset := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("POP_JUMP_IF_NOT_ERROR %d\n", offset))
			pc += 4

		case OpJump:
			offset := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("JUMP %d\n", offset))
			pc += 4

		case OpAbsoluteJump:
			offset := ReadInt(code.Code, pc)
			builder.WriteString(fmt.Sprintf("ABSOLUTE_JUMP %d\n", offset))
			pc += 4

		case OpDupTop:
			builder.WriteString("DUP_TOP\n")

		case OpNoOp:
			builder.WriteString("NO_OP\n")

		case OpPopTop:
			builder.WriteString("POP_TOP\n")

		case OpReturn:
			builder.WriteString("RETURN\n")

		default:
			builder.WriteString(fmt.Sprintf("UNKNOWN_OPCODE %d\n", opcode))
		}
	}

	return strings.TrimSpace(builder.String())
}
