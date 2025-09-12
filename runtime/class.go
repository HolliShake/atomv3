package runtime

type AtomClass struct {
	Name  string
	Base  *AtomValue // AtomClass
	Proto *AtomValue // AtomObject
}

func NewAtomClass(name string, base, proto *AtomValue) *AtomClass {
	return &AtomClass{
		Name:  name,
		Base:  base,
		Proto: proto,
	}
}
