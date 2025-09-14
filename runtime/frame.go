package runtime

type AtomCallFrame struct {
	Fn      *AtomValue // Function
	Env     *AtomEnv
	Ip      int
	Stack   int
	Promise *AtomValue // For delayed task
}

func NewAtomCallFrame(fn *AtomValue, env *AtomEnv, ip int) *AtomCallFrame {
	return &AtomCallFrame{
		Fn:      fn,
		Env:     env,
		Ip:      ip,
		Stack:   0,
		Promise: nil,
	}
}
