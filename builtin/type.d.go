package builtin

import runtime "dev.runtime"

type NativeModule struct {
	Functions map[string]*runtime.AtomValue
}
