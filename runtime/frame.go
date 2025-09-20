package runtime

type AtomCallFrame struct {
	Caller  *AtomCallFrame // Caller
	Fn      *AtomValue     // Function
	Pc      int            // Basically thesame with Ip, difference is that it only increments 1
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
