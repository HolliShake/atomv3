# Atom Programming Language

Atom is a custom programming language implemented in Go, featuring a complete compiler and interpreter with support for functions, variables, control flow, and various data types.

## Features

- **Dynamic Typing**: Supports integers, floats, strings, booleans, arrays, and functions
- **Function Support**: First-class functions with parameters and return statements
- **Variable Declarations**: `var`, `const`, and `local` variable declarations
- **Control Flow**: `if-else` statements with proper scoping
- **Arithmetic Operations**: Full set of mathematical and bitwise operations
- **Array Support**: Array literals and operations
- **Error Handling**: Comprehensive error reporting with source location
- **Garbage Collection**: Basic memory management with GC roots

## Language Syntax

### Data Types

Atom supports the following primitive types:
- **Integers**: `42`, `-17`, `0`
- **Floats**: `3.14`, `-2.5`, `0.0`
- **Strings**: `"hello"`, `"world"`
- **Booleans**: `true`, `false`
- **Null**: `null`
- **Arrays**: `[1, 2, 3, 4, 5]`
- **Functions**: First-class function objects

### Variable Declarations

```atom
// Global variables
var x = 1;
var y = 2, z = 3;

// Constants
const PI = 3.14159;
const MAX_SIZE = 100;

// Local variables (function-scoped)
local temp = x + y;
```

### Functions

```atom
// Function declaration
func add(a, b) {
    return a + b;
}

// Function with multiple parameters
func greet(name, age) {
    return "Hello " + name + ", you are " + age + " years old";
}

// Recursive function
func factorial(n) {
    if (n == 0) {
        return 1;
    }
    return n * factorial(n - 1);
}
```

### Control Flow

```atom
// If-else statements
if (x > 0) {
    var result = "positive";
} else {
    var result = "negative or zero";
}

// Nested if statements
if (age >= 18) {
    if (hasLicense) {
        var status = "can drive";
    } else {
        var status = "needs license";
    }
} else {
    var status = "too young";
}
```

### Arithmetic Operations

```atom
// Basic arithmetic
var sum = 2 + 3;        // 5
var diff = 10 - 4;      // 6
var product = 3 * 4;    // 12
var quotient = 15 / 3;  // 5
var remainder = 17 % 5; // 2

// Bitwise operations
var and = 5 & 3;        // 1
var or = 5 | 3;         // 7
var xor = 5 ^ 3;        // 6
var leftShift = 4 << 1; // 8
var rightShift = 8 >> 1; // 4

// Comparison operations
var isEqual = (5 == 5);     // true
var isNotEqual = (5 != 3);  // true
var isGreater = (10 > 5);   // true
var isLess = (3 < 7);       // true
var isGreaterEqual = (5 >= 5); // true
var isLessEqual = (4 <= 6);    // true

// Logical operations
var logicalAnd = (true && false); // false
var logicalOr = (true || false);  // true
```

### Arrays

```atom
// Array literals
var numbers = [1, 2, 3, 4, 5];
var mixed = [1, "hello", true, 3.14];
var empty = [];

// Arrays with function calls
var results = [factorial(3), factorial(4), factorial(5)];
```

### Expression Statements

```atom
// Standalone expressions
2 + 3 * 4;
add(10, 20);
factorial(5);

// Complex expressions
(2 + 3) * (4 - 1);
add(multiply(2, 3), divide(10, 2));
```

## Example Programs

### Hello World
```atom
var message = "Hello, World!";
message;
```

### Fibonacci Sequence
```atom
func fibonacci(n) {
    if (n <= 1) {
        return n;
    }
    return fibonacci(n - 1) + fibonacci(n - 2);
}

const n = 10;
fibonacci(n);
```

### Array Operations
```atom
func sumArray(arr) {
    var total = 0;
    // Note: Array iteration would need to be implemented
    return total;
}

var numbers = [1, 2, 3, 4, 5];
sumArray(numbers);
```

### Complex Example
```atom
var x = 1;

func add(a, b) {
   return a + b;
}

2 + 23 / 23 - 2;

3 + 2 + add(12, 2);

if (0 || 0) {
    2 + 2;
} else {
    3 + 3;
    local x = 69420;
    local y = x + 1;
    x;
    y;
}

1 && 2;

func fact(n) {
    if (n == 0) {
        return 1;
    }
    return n * fact(n - 1);
}

const n = 3 + 2;
fact(n);

[1, 2, 3, 4, 5, 6, fact(n)];
```

## Architecture

### Compiler Pipeline

1. **Tokenizer** (`atom_tokenizer.go`): Converts source code into tokens
2. **Parser** (`atom_parser.go`): Builds Abstract Syntax Tree (AST) from tokens
3. **Compiler** (`atom_compile.go`): Compiles AST to bytecode
4. **Interpreter** (`atom_interpreter.go`): Executes bytecode

### Runtime System

- **Stack-based VM**: Uses evaluation stack for expression evaluation
- **Frame Management**: Handles function calls and local scopes
- **Memory Management**: Basic garbage collection with GC roots
- **Type System**: Dynamic typing with runtime type checking

### Bytecode Operations

The VM supports the following opcodes:
- **Load Operations**: `OpLoadInt`, `OpLoadNum`, `OpLoadStr`, `OpLoadArray`
- **Arithmetic**: `OpAdd`, `OpSub`, `OpMul`, `OpDiv`, `OpMod`
- **Bitwise**: `OpAnd`, `OpOr`, `OpXor`, `OpShl`, `OpShr`
- **Comparison**: `OpCmpLt`, `OpCmpGt`, `OpCmpEq`, `OpCmpNe`
- **Control Flow**: `OpJump`, `OpJumpIfFalse`, `OpJumpIfTrue`
- **Function Calls**: `OpCall`, `OpReturn`
- **Variable Access**: `OpLoadLocal`, `OpStoreLocal`, `OpLoadCapture`

## Building and Running

### Prerequisites
- Go 1.19 or later

### Build
```bash
# Build the compiler/interpreter
cd app
go build -o atom .

# Or run directly
go run *.go <file.atom>
```

### Usage
```bash
# Run an Atom program
./atom test.atom

# Or with go run
go run *.go test.atom
```

## Project Structure

```
atom/
├── app/                    # Main application and compiler
│   ├── main.go            # Entry point
│   ├── atom_tokenizer.go  # Lexical analysis
│   ├── atom_parser.go     # Syntax analysis
│   ├── atom_compile.go    # Code generation
│   ├── atom_ast.go        # AST definitions
│   ├── atom_error.go      # Error handling
│   └── test.atom          # Example program
├── runtime/               # Runtime system
│   ├── atom_interpreter.go # VM implementation
│   ├── atom_operation.go   # Operation implementations
│   ├── atom_value.go      # Value system
│   ├── atom_opcode.go     # Bytecode definitions
│   └── atom_stack.go      # Stack implementation
└── builtin/               # Built-in types and functions
```

## Language Limitations

- No loops (`for`, `while`) - only recursive functions
- No object-oriented programming
- No modules or imports
- No exception handling
- Limited array operations
- No string manipulation functions
- No I/O operations

## Future Enhancements

- Loop constructs (`for`, `while`)
- Object-oriented features
- Module system
- Standard library
- Better error messages
- Optimizations
- REPL (Read-Eval-Print Loop)

## License

This project is licensed under the terms specified in the LICENSE file.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests to improve the language implementation.
