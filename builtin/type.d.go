package builtin

import runtime "dev.runtime"

type NativeFunction func(intereter *runtime.AtomInterpreter, argc int)
