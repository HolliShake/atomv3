package runtime

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

func (s *AtomScheduler) Running(frame *AtomCallFrame) {
	if !frame.Fn.Value.(*AtomCode).Async {
		return
	}
	if frame.State == ExecIdle {
		// From idle to running
		frame.State = ExecRunning
		// Create a promise if ever the function meets an await,
		// notice func (s *AtomScheduler) Await(frame *AtomCallFrame);
		frame.Promise = NewAtomValuePromise(PromiseStatePending, nil)
	}
}

func (s *AtomScheduler) Await(frame *AtomCallFrame) (suspend bool) {
	t := frame.Stack.Pop()
	p := t.Value.(*AtomPromise)
	if p.State == PromiseStateFulfilled {
		frame.State = ExecRunning
		// push the awaited value to the current frame's Stack
		frame.Stack.Push(
			p.Value,
		)
		return false
	} else {
		frame.State = ExecAwaiting
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
	defer func() {
		for _, cell := range frame.Fn.Value.(*AtomCode).Locals {
			if !cell.Captured {
				cell.Value = nil
			}
		}
	}()

	// Handle synchronous functions
	if !frame.Fn.Value.(*AtomCode).Async {
		if frame.Caller != nil {
			frame.Caller.Stack.Push(
				frame.Stack.Pop(),
			)
		}

		frame.Stack.Clear()
		frame.Promise = nil
		frame.State = ExecIdle
		return
	}

	// Handle asynchronous functions
	frame.State = ExecCompleted

	// Get the promise and fulfill it with the return value
	promise := frame.Promise.Value.(*AtomPromise)
	promise.State = PromiseStateFulfilled
	promise.Value = frame.Stack.Pop()

	// Push the fulfilled promise to caller's stack
	if frame.Caller != nil && !frame.Caller.Fn.Value.(*AtomCode).Async {
		frame.Caller.Stack.Push(
			promise.Value,
		)
	} else {
		frame.Caller.Stack.Push(
			frame.Promise,
		)
	}

	// Clean up frame state
	frame.Stack.Clear()
	frame.Promise = nil
	frame.State = ExecIdle
}

func (s *AtomScheduler) Run() {
	for len(s.MicroTask) > 0 {
		task := s.MicroTask[0]
		s.MicroTask = s.MicroTask[1:]
		s.Interpreter.ExecuteFrame(task)
	}
}
