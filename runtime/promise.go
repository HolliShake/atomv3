package runtime

type PromiseState int

const (
	PromiseStatePending PromiseState = iota
	PromiseStateFulfilled
	PromiseStateRejected
)

type AtomPromise struct {
	State PromiseState
	Value *AtomValue
}

func NewAtomPromise(state PromiseState, value *AtomValue) *AtomPromise {
	return &AtomPromise{
		State: state,
		Value: value,
	}
}

func (p *AtomPromise) HashValue() int {
	hash := 0

	// Hash the promise state
	hash = hash*31 + int(p.State)

	// Hash the promise value if it exists
	if p.Value != nil {
		hash = hash*31 + p.Value.HashValue()
	}

	return hash
}

func (p *AtomPromise) IsFulfilled() bool {
	return p.State == PromiseStateFulfilled
}

func (p PromiseState) String() string {
	switch p {
	case PromiseStatePending:
		return "<pending>"
	case PromiseStateFulfilled:
		return "<fulfilled>"
	case PromiseStateRejected:
		return "<rejected>"
	}
	return "unknown"
}
