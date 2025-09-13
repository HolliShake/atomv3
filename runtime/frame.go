package runtime

type AtomCallFrame struct {
	Fn  *AtomValue // Function
	Env *AtomEnv
	Ip  int
}

func NewAtomCallFrame(fn *AtomValue, env *AtomEnv, ip int) *AtomCallFrame {
	return &AtomCallFrame{
		Fn:  fn,
		Env: env,
		Ip:  ip,
	}
}
