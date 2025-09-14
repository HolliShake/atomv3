package runtime

import "fmt"

type ExecutionState int

const (
	ExecutionStateIdle ExecutionState = iota
	ExecutionStateRunning
	ExecutionStateAwaiting
	ExecutionStateCompleted
)

type Transition struct {
	From      ExecutionState
	To        ExecutionState
	Condition func(*AtomInterpreter, *AtomCallFrame) bool
	Action    func(*AtomInterpreter, *AtomCallFrame) error
}

func Wait(interpreter *AtomInterpreter, caller *AtomCallFrame) {
	waitingFor := interpreter.popp()
	// if waitingFor == nil {
	// 	panic(fmt.Sprintf("%s is waiting for nil", caller.Fn.Value.(*AtomCode).Name))
	// }
	caller.State = ExecutionStateAwaiting
	caller.Value = waitingFor
	interpreter.MicroTask.Enqueue(caller)
	// Don't push a promise here - the caller should get the actual resolved value
}

func ExecuteTransition(interpreter *AtomInterpreter, frame *AtomCallFrame) {
	if !frame.Fn.Value.(*AtomCode).Async {
		return
	}

	// Don't execute transitions on already completed frames
	if frame.State == ExecutionStateCompleted {
		return
	}

	switch frame.State {
	case ExecutionStateIdle: // -> run
		{
			frame.State = ExecutionStateRunning
			// Create a pending promise that will be resolved when the function completes
			frame.Value = NewAtomValuePromise(PromiseStatePending, nil)
			// Immediately push the promise to the stack so callers can await it
			interpreter.pushVal(frame.Value)
		}
	case ExecutionStateRunning: // -> completed
		{
			frame.State = ExecutionStateCompleted

			// For async functions, resolve the pending promise
			code := frame.Fn.Value.(*AtomCode)
			if code.Async {
				if CheckType(frame.Value, AtomTypePromise) {
					// This is the pending promise we created earlier, resolve it
					promise := frame.Value.Value.(*AtomPromise)
					// Get the return value from the stack
					if interpreter.EvaluationStack.Len() > 0 {
						returnValue := interpreter.popp()
						promise.State = PromiseStateFulfilled
						promise.Value = returnValue
					} else {
						// No return value, resolve with null
						promise.State = PromiseStateFulfilled
						promise.Value = interpreter.State.NullValue
					}
				} else {
					// This shouldn't happen for async functions
					panic("Async function frame.Value is not a promise")
				}
			} else {
				// For non-async functions, just push the value
				if frame.Value != nil {
					interpreter.pushVal(frame.Value)
				}
			}
		}
	case ExecutionStateAwaiting: // -> completed
		{
			if frame.Value == nil {
				fmt.Println(">>", frame.Fn.Value.(*AtomCode).Name, "Completed", frame.Value)
				panic("frame.Value is nil in ExecutionStateAwaiting")
			}

			// Check if the awaited value is a fulfilled promise
			if CheckType(frame.Value, AtomTypePromise) {
				promise := frame.Value.Value.(*AtomPromise)
				if promise.State == PromiseStateFulfilled {
					// The await is complete, continue execution
					frame.State = ExecutionStateRunning
					// Push the resolved value to the stack for the await expression
					interpreter.pushVal(promise.Value)
				} else {
					// Still pending, keep waiting
					frame.State = ExecutionStateAwaiting
				}
			} else {
				// Not a promise, directly use the value
				frame.State = ExecutionStateRunning
				interpreter.pushVal(frame.Value)
			}
		}
	case ExecutionStateCompleted:
		{
			panic("Attempting to execute transition on already completed frame")
		}
	default:
		panic("Unknown execution state encountered")
	}
}

func ExecuteMicroTask(interpreter *AtomInterpreter) {
	for !interpreter.MicroTask.IsEmpty() {
		top, _ := interpreter.MicroTask.Dequeue()
		// Only execute the frame if it's not completed
		if top.State != ExecutionStateCompleted {
			// Check if the frame is awaiting and the awaited value is ready
			if top.State == ExecutionStateAwaiting {
				// Check if the awaited value is ready
				if top.Value != nil {
					if CheckType(top.Value, AtomTypePromise) {
						promise := top.Value.Value.(*AtomPromise)
						if promise.State == PromiseStateFulfilled {
							// The promise is fulfilled, we can continue execution
							interpreter.executeFrame(top)
						} else {
							// Still pending, re-enqueue
							interpreter.MicroTask.Enqueue(top)
						}
					} else {
						// Not a promise, we can continue execution
						interpreter.executeFrame(top)
					}
				} else {
					// No value to wait for, continue execution
					interpreter.executeFrame(top)
				}
			} else {
				// Not awaiting, execute normally
				interpreter.executeFrame(top)
			}
		}
	}
}
