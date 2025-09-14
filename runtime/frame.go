package runtime

type AtomCallFrame struct {
	Fn    *AtomValue     // Function
	Env   *AtomEnv       // Environment
	Ip    int            // Instruction pointer
	State ExecutionState // For async/await
	Value *AtomValue     // For delayed task
}

func NewAtomCallFrame(fn *AtomValue, env *AtomEnv, ip int) *AtomCallFrame {
	return &AtomCallFrame{
		Fn:    fn,
		Env:   env,
		Ip:    ip,
		State: ExecutionStateIdle,
		Value: nil,
	}
}
