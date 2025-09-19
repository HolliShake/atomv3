package runtime

import "fmt"

type ExecutionState int

const (
	ExecIdle ExecutionState = iota
	ExecAwaiting
	ExecRunning
	ExecCompleted
)

type AtomScheduler struct {
	Interpreter *AtomInterpreter
	MicroTask   []*AtomCallFrame
}

func NewAtomScheduler(interpreter *AtomInterpreter) *AtomScheduler {
	return &AtomScheduler{
		Interpreter: interpreter,
		MicroTask:   []*AtomCallFrame{},
	}
}

func (s *AtomScheduler) MoveNextEvent(frame *AtomCallFrame) {
	if !frame.Fn.Value.(*AtomCode).Async {
		return
	}
	switch frame.State {
	case ExecIdle:
		// From idle to running
		frame.State = ExecRunning
		// Create a promise if ever the function meets an await,
		// notice func (s *AtomScheduler) Await(frame *AtomCallFrame);
		frame.Promise = NewAtomValuePromise(PromiseStatePending, nil)

	case ExecAwaiting:
		// From awaiting to running
		// frame.Stack.Push(frame.Promise)

	case ExecRunning:
		// From running to completed
		frame.State = ExecCompleted

	default:
		panic(fmt.Sprintf("Unknown execution state: %d", frame.State))
	}
}

func (s *AtomScheduler) Await(frame *AtomCallFrame) (suspend bool) {
	t := frame.Stack.Pop()
	p := t.Value.(*AtomPromise)
	if p.State == PromiseStateFulfilled {
		// push the awaited value to the current frame's Stack
		frame.Stack.Push(p.Value)
		return false
	} else {
		// Push the current frame's promise to it's caller
		frame.Caller.Stack.Push(
			frame.Promise,
		)
		s.MicroTask = append(s.MicroTask, frame)
	}
	// Suspend process
	return true
}

func (s *AtomScheduler) Resolve(frame *AtomCallFrame) {
	if !frame.Fn.Value.(*AtomCode).Async {
		return
	}
	frame.State = ExecCompleted
	frame.Promise.Value.(*AtomPromise).State = PromiseStateFulfilled
	frame.Promise.Value.(*AtomPromise).Value = frame.Stack.Pop()
	frame.Caller.Stack.Push(frame.Promise)
}

func (s *AtomScheduler) Run() {
	for len(s.MicroTask) > 0 {
		task := s.MicroTask[0]
		s.MicroTask = s.MicroTask[1:]
		s.Interpreter.ExecuteFrame(task)
	}
}
