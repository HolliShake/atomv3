package runtime

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func getGin(obj *AtomValue) *gin.Engine {
	return obj.Obj.(*AtomClassInstance).Property.Obj.(*gin.Engine)
}

func getParams(c *gin.Context) *AtomValue {
	params := NewAtomObject(map[string]*AtomValue{})
	for _, param := range c.Params {
		params.Elements[param.Key] = NewAtomValueStr(param.Value)
	}
	return NewAtomGenericValue(AtomTypeObj, params)
}

func getBody(c *gin.Context) *AtomValue {
	body := map[string]any{}
	if err := c.BindJSON(&body); err != nil {
		return NewAtomValueError(err.Error())
	}
	return ToAtomObject(body)
}

func getStatus(obj *AtomValue) int {
	defaultStatus := 422
	if obj == nil {
		return defaultStatus
	}

	// Helper to validate and return status
	validateStatus := func(status int) int {
		if status >= 100 && status < 600 {
			return status
		}
		return defaultStatus
	}

	// Helper to get status from AtomValue
	getStatusFromValue := func(val *AtomValue) int {
		switch val.Type {
		case AtomTypeInt:
			return validateStatus(int(val.I32))
		case AtomTypeNum:
			return validateStatus(int(val.F64))
		default:
			return defaultStatus
		}
	}

	// Helper to get status from object's "status" field
	getStatusFromObject := func(atomObj *AtomObject) int {
		if statusVal, exists := atomObj.Elements["status"]; exists {
			return getStatusFromValue(statusVal)
		}
		return defaultStatus
	}

	switch obj.Type {
	case AtomTypeInt:
		return validateStatus(int(obj.I32))
	case AtomTypeNum:
		return validateStatus(int(obj.F64))
	case AtomTypeObj:
		return getStatusFromObject(obj.Obj.(*AtomObject))
	case AtomTypeClass:
		atomClass := obj.Obj.(*AtomClass)
		if atomClass.Proto != nil && atomClass.Proto.Type == AtomTypeObj {
			return getStatusFromObject(atomClass.Proto.Obj.(*AtomObject))
		}
		return defaultStatus
	case AtomTypeClassInstance:
		classInstance := obj.Obj.(*AtomClassInstance)
		if classInstance.Property != nil && classInstance.Property.Type == AtomTypeObj {
			return getStatusFromObject(classInstance.Property.Obj.(*AtomObject))
		}
		return defaultStatus
	default:
		return defaultStatus
	}
}

func builtin_init_gin() *AtomValue {
	var class = NewAtomGenericValue(
		AtomTypeClass,
		NewAtomClass("Gin", nil, NewAtomGenericValue(
			AtomTypeObj,
			NewAtomObject(map[string]*AtomValue{}),
		)),
	)

	var protoType = class.Obj.(*AtomClass).Proto.Obj.(*AtomObject)

	// func init() -> gin.Default();
	var constructor = NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("init", Variadict, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
			CleanupStack(frame, argc)
			// Push this
			this := NewAtomGenericValue(
				AtomTypeClassInstance,
				// prototype, property
				NewAtomClassInstance(class, NewAtomGenericValue(AtomTypeObj, gin.New())),
			)
			frame.Stack.Push(this)
		}),
	)

	// func get(path, callback) -> gin.GET
	var get = NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("get", 3, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
			if argc != 3 {
				CleanupStack(frame, argc)
				message := FormatError(frame, "get expects 2 arguments")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			this := frame.Stack.GetOffset(argc, 0)     // class instance
			path := frame.Stack.GetOffset(argc, 1)     // path
			callback := frame.Stack.GetOffset(argc, 2) // callback

			if !CheckType(this, AtomTypeClassInstance) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "get expects a Gin instance")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			if !CheckType(path, AtomTypeStr) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "get expects a path string")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			if !CheckType(callback, AtomTypeFunc) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "get expects a callback user defined function")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			CleanupStack(frame, argc)

			ginInstance := getGin(this)
			ginInstance.GET(path.String(), func(c *gin.Context) {
				// Create an object value for params
				objValue := getParams(c)
				// Push as argument
				frame.Stack.Push(objValue)
				// Call
				DoCall(interpreter, frame, callback, 1)
				result := frame.Stack.Pop()
				// Response
				c.JSON(getStatus(result), SerializeObject(result))
			})

			frame.Stack.Push(this)
		}),
	)

	// func post(path, callback) -> gin.POST
	var post = NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("post", 3, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
			if argc != 3 {
				CleanupStack(frame, argc)
				message := FormatError(frame, "post expects 2 arguments")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			this := frame.Stack.GetOffset(argc, 0)     // class instance
			path := frame.Stack.GetOffset(argc, 1)     // path
			callback := frame.Stack.GetOffset(argc, 2) // callback

			if !CheckType(this, AtomTypeClassInstance) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "post expects a Gin instance")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			if !CheckType(path, AtomTypeStr) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "post expects a path string")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			if !CheckType(callback, AtomTypeFunc) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "post expects a callback user defined function")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			CleanupStack(frame, argc)

			ginInstance := getGin(this)
			ginInstance.POST(path.String(), func(c *gin.Context) {
				// Create an object value for params
				objValue := NewAtomGenericValue(AtomTypeObj, NewAtomObject(map[string]*AtomValue{
					"body":   getBody(c),
					"params": getParams(c),
				}))
				// Push as argument
				frame.Stack.Push(objValue)
				// Call
				DoCall(interpreter, frame, callback, 1)
				result := frame.Stack.Pop()
				// Response
				c.JSON(getStatus(result), SerializeObject(result))
			})

			frame.Stack.Push(this)
		}),
	)

	// func put(path, callback) -> gin.PUT
	var put = NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("put", 3, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
			if argc != 3 {
				CleanupStack(frame, argc)
				message := FormatError(frame, "put expects 2 arguments")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			this := frame.Stack.GetOffset(argc, 0)     // class instance
			path := frame.Stack.GetOffset(argc, 1)     // path
			callback := frame.Stack.GetOffset(argc, 2) // callback

			if !CheckType(this, AtomTypeClassInstance) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "put expects a Gin instance")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			if !CheckType(path, AtomTypeStr) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "put expects a path string")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			if !CheckType(callback, AtomTypeFunc) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "put expects a callback user defined function")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			CleanupStack(frame, argc)

			ginInstance := getGin(this)
			ginInstance.PUT(path.String(), func(c *gin.Context) {
				// Create an object value for params
				objValue := NewAtomGenericValue(AtomTypeObj, NewAtomObject(map[string]*AtomValue{
					"body":   getBody(c),
					"params": getParams(c),
				}))
				// Push as argument
				frame.Stack.Push(objValue)
				// Call
				DoCall(interpreter, frame, callback, 1)
				result := frame.Stack.Pop()
				// Response
				c.JSON(getStatus(result), SerializeObject(result))
			})

			frame.Stack.Push(this)
		}),
	)

	// func patch(path, callback) -> gin.PATCH
	var patch = NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("patch", 3, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
			if argc != 3 {
				CleanupStack(frame, argc)
				message := FormatError(frame, "patch expects 2 arguments")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			this := frame.Stack.GetOffset(argc, 0)     // class instance
			path := frame.Stack.GetOffset(argc, 1)     // path
			callback := frame.Stack.GetOffset(argc, 2) // callback

			if !CheckType(this, AtomTypeClassInstance) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "patch expects a Gin instance")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			if !CheckType(path, AtomTypeStr) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "patch expects a path string")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			if !CheckType(callback, AtomTypeFunc) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "patch expects a callback user defined function")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			CleanupStack(frame, argc)

			ginInstance := getGin(this)
			ginInstance.PATCH(path.String(), func(c *gin.Context) {
				// Create an object value for params
				objValue := NewAtomGenericValue(AtomTypeObj, NewAtomObject(map[string]*AtomValue{
					"body":   getBody(c),
					"params": getParams(c),
				}))
				// Push as argument
				frame.Stack.Push(objValue)
				// Call
				DoCall(interpreter, frame, callback, 1)
				result := frame.Stack.Pop()
				// Response
				c.JSON(getStatus(result), SerializeObject(result))
			})

			frame.Stack.Push(this)
		}),
	)

	// func delete(path, callback) -> gin.DELETE
	var delete = NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("delete", 3, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
			if argc != 3 {
				CleanupStack(frame, argc)
				message := FormatError(frame, "delete expects 2 arguments")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			this := frame.Stack.GetOffset(argc, 0)     // class instance
			path := frame.Stack.GetOffset(argc, 1)     // path
			callback := frame.Stack.GetOffset(argc, 2) // callback

			if !CheckType(this, AtomTypeClassInstance) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "delete expects a Gin instance")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			if !CheckType(path, AtomTypeStr) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "delete expects a path string")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			if !CheckType(callback, AtomTypeFunc) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "delete expects a callback user defined function")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			CleanupStack(frame, argc)

			ginInstance := getGin(this)
			ginInstance.DELETE(path.String(), func(c *gin.Context) {
				// Create an object value for params
				objValue := getParams(c)
				// Push as argument
				frame.Stack.Push(objValue)
				// Call
				DoCall(interpreter, frame, callback, 1)
				result := frame.Stack.Pop()
				// Response
				c.JSON(getStatus(result), SerializeObject(result))
			})

			frame.Stack.Push(this)
		}),
	)

	// func serve() -> gin.Run();
	var serve = NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("serve", 2, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
			if argc != 2 {
				CleanupStack(frame, argc)
				message := FormatError(frame, "serve expects 0 arguments")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			this := frame.Stack.GetOffset(argc, 0) // class instance
			port := frame.Stack.GetOffset(argc, 1) // port number

			if !CheckType(this, AtomTypeClassInstance) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "serve expects a Gin instance")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			if !IsNumberType(port) {
				CleanupStack(frame, argc)
				message := FormatError(frame, "serve expects a port number")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			// Check if valid port range
			portNum := CoerceToInt(port)
			if portNum < 1 || portNum > 65535 {
				CleanupStack(frame, argc)
				message := FormatError(frame, "serve expects a valid port number")
				frame.Stack.Push(NewAtomValueError(message))
				return
			}

			gin := getGin(this)
			gin.Run(fmt.Sprintf(":%d", portNum))
			frame.Stack.Push(interpreter.State.NullValue)
		}),
	)

	protoType.Set("init", constructor)
	protoType.Set("get", get)
	protoType.Set("post", post)
	protoType.Set("put", put)
	protoType.Set("patch", patch)
	protoType.Set("delete", delete)
	protoType.Set("serve", serve)

	return class
}

func gin_created(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		message := FormatError(frame, fmt.Sprintf("created expected 1 argument, got %d", argc))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	data := frame.Stack.Pop()
	objValue := NewAtomGenericValue(AtomTypeObj, NewAtomObject(map[string]*AtomValue{
		"status": NewAtomValueInt(201),
		"data":   data,
	}))

	frame.Stack.Push(objValue)
}

func gin_ok(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		message := FormatError(frame, fmt.Sprintf("ok expected 1 argument, got %d", argc))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	data := frame.Stack.Pop()
	objValue := NewAtomGenericValue(AtomTypeObj, NewAtomObject(map[string]*AtomValue{
		"status": NewAtomValueInt(200),
		"data":   data,
	}))

	frame.Stack.Push(objValue)
}

func gin_badRequest(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		message := FormatError(frame, fmt.Sprintf("badRequest expected 1 argument, got %d", argc))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	data := frame.Stack.Pop()
	objValue := NewAtomGenericValue(AtomTypeObj, NewAtomObject(map[string]*AtomValue{
		"status": NewAtomValueInt(400),
		"data":   data,
	}))

	frame.Stack.Push(objValue)
}

func gin_unauthorized(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		message := FormatError(frame, fmt.Sprintf("unauthorized expected 1 argument, got %d", argc))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	data := frame.Stack.Pop()
	objValue := NewAtomGenericValue(AtomTypeObj, NewAtomObject(map[string]*AtomValue{
		"status": NewAtomValueInt(401),
		"data":   data,
	}))

	frame.Stack.Push(objValue)
}

func gin_forbidden(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		message := FormatError(frame, fmt.Sprintf("forbidden expected 1 argument, got %d", argc))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	data := frame.Stack.Pop()
	objValue := NewAtomGenericValue(AtomTypeObj, NewAtomObject(map[string]*AtomValue{
		"status": NewAtomValueInt(403),
		"data":   data,
	}))

	frame.Stack.Push(objValue)
}

func gin_notFound(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		message := FormatError(frame, fmt.Sprintf("notFound expected 1 argument, got %d", argc))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	data := frame.Stack.Pop()
	objValue := NewAtomGenericValue(AtomTypeObj, NewAtomObject(map[string]*AtomValue{
		"status": NewAtomValueInt(404),
		"data":   data,
	}))

	frame.Stack.Push(objValue)
}

func gin_internalServerError(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		message := FormatError(frame, fmt.Sprintf("internalServerError expected 1 argument, got %d", argc))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	data := frame.Stack.Pop()
	objValue := NewAtomGenericValue(AtomTypeObj, NewAtomObject(map[string]*AtomValue{
		"status": NewAtomValueInt(500),
		"data":   data,
	}))

	frame.Stack.Push(objValue)
}

func gin_response(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 2 {
		CleanupStack(frame, argc)
		message := FormatError(frame, fmt.Sprintf("internalServerError expected 1 argument, got %d", argc))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	status := frame.Stack.GetOffset(argc, 0)
	data := frame.Stack.GetOffset(argc, 1)

	if !IsNumberType(status) {
		CleanupStack(frame, argc)
		message := FormatError(frame, fmt.Sprintf("status code must be a type of int, got %s", GetTypeString(status)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	objValue := NewAtomGenericValue(AtomTypeObj, NewAtomObject(map[string]*AtomValue{
		"status": NewAtomValueInt(int(CoerceToInt(status))),
		"data":   data,
	}))

	frame.Stack.Push(objValue)
}

var EXPORT_GIN = map[string]*AtomValue{
	"Gin": builtin_init_gin(),
	"created": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("created", 1, gin_created),
	),
	"ok": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("ok", 1, gin_ok),
	),
	"badRequest": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("badRequest", 1, gin_badRequest),
	),
	"unauthorized": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("unauthorized", 1, gin_unauthorized),
	),
	"forbidden": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("forbidden", 1, gin_forbidden),
	),
	"notFound": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("notFound", 1, gin_notFound),
	),
	"internalServerError": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("internalServerError", 1, gin_internalServerError),
	),
	// Generic
	"response": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("response", 2, gin_response),
	),
}
