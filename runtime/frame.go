package runtime

type AtomCallFrame struct {
	Caller  *AtomCallFrame // Caller
	Fn      *AtomValue     // Function
	Ip      int            // Instruction pointer
	Stack   *AtomStack     // EvaluationStack
	Promise *AtomValue     // Promise
	State   ExecutionState // For async/await
}

func NewAtomCallFrame(caller *AtomCallFrame, fn *AtomValue, ip int) *AtomCallFrame {
	return &AtomCallFrame{
		Caller:  caller,
		Fn:      fn,
		Ip:      ip,
		Stack:   NewAtomStack(),
		Promise: nil,
	}
}
