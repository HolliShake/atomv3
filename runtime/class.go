package runtime

type AtomClass struct {
	Base  *AtomValue // AtomClass
	Proto *AtomValue // AtomObject
}

func NewAtomClass(base, proto *AtomValue) *AtomClass {
	return &AtomClass{
		Base:  nil,
		Proto: nil,
	}
}
