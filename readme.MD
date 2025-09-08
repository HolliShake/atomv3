# Atom Programming Language

Atom is a custom programming language implemented in Go, featuring a complete compiler and interpreter with support for functions, variables, control flow, objects, arrays, and various data types. It uses a stack-based virtual machine for execution.

## Features

- **Dynamic Typing**: Supports integers, floats, strings, booleans, arrays, objects, and functions
- **Function Support**: First-class functions with parameters and return statements
- **Variable Declarations**: `var`, `const`, and `local` variable declarations with proper scoping
- **Control Flow**: `if-else`, `switch` expressions, `while`, and `do-while`
- **Arithmetic Operations**: Full set of mathematical, bitwise, and logical operations
- **Assignment Operators**: Compound assignment operators (`+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=`)
- **Array Support**: Array literals, indexing, and element access
- **Object Support**: Object literals with key-value pairs, property access, and method calls
- **Imports & Stdlib**: Import from `atom:std` (e.g., `print`, `println`)
- **Stack-based VM**: Efficient bytecode execution with evaluation stack
- **Error Handling**: Comprehensive error reporting with source location
- **Memory Management**: Basic garbage collection with GC roots

## Language Syntax

### Data Types

Atom supports the following primitive types:
- **Integers**: `42`, `-17`, `0` (32-bit signed integers)
- **Floats**: `3.14`, `-2.5`, `0.0` (64-bit floating point numbers)
- **Strings**: `"hello"`, `"world"` (Unicode strings with escape sequences)
- **Booleans**: `true`, `false`
- **Null**: `null`
- **Arrays**: `[1, 2, 3, 4, 5]` (dynamic arrays with indexing)
- **Objects**: `{key: value, name: "test"}` (key-value pairs with property access)
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

// Array indexing
var first = numbers[0];        // Access first element
numbers[1] = 42;              // Modify second element
```

### Objects

```atom
// Object literals
var person = {
    name: "Alice",
    age: 30,
    greet: function() {
        return "Hello, I'm " + this.name;
    }
};

// Property access
var name = person.name;        // Dot notation
var age = person["age"];       // Bracket notation

// Property assignment
person.name = "Bob";
person["age"] = 31;

// Method calls
var greeting = person.greet();
```

### Assignment Operators

```atom
var x = 10;

// Compound assignment operators
x += 5;    // x = x + 5 (15)
x -= 3;    // x = x - 3 (12)
x *= 2;    // x = x * 2 (24)
x /= 4;    // x = x / 4 (6)
x %= 5;    // x = x % 5 (1)

// Bitwise assignment operators
x &= 3;    // x = x & 3
x |= 7;    // x = x | 7
x ^= 4;    // x = x ^ 4
x <<= 2;   // x = x << 2
x >>= 1;   // x = x >> 1
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

### Imports and Standard Library I/O

```atom
// Import selected symbols from the standard library
import [ print, println ] from "atom:std";

println("Hello, World!", 1, 2, 3);
print("Wheels on the bus!", std, "\n");

const Log = println; // aliasing
Log("Hello, World!");
```

### Switch and Conditional Expressions

```atom
// Switch as an expression
const value = (5) switch {
    case(1, 2) => 200
    case(3, 4) => 300
    default    => 400
};

println(">>", value);

// If as an expression
const cond = 
    if (3 < 2)
        2 + 2
    else if (3 < 2)
        3 + 3
    else
        4 + 4
    ;

println(cond);
```

### Loops: while and do-while

```atom
var i = 0;
while (1 && i < 10) {
    println(i);
    i += 1;
}

do {
    println(i);
    i += 1;
} while (i < 20);
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

### Complex Example (from test.atom)
```atom
var x = 1;

func add(a, b) {
   return a + b;
}

2 + 23 / 23 - 2;

3 + 2 + add(12, 2);

if (0 || 0)
    2 + 2;
else {
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

[1,2,3,4,5,6, fact(n), "Nah!!"];

var obj = {
    getName: "Hello",
    add: add,
    val: 0
};

obj;
obj.getName;

obj.val += 2;
obj["val"] += 2;

obj;

var xh = 12;

xh *= 2;
xh = 100;

obj.getName = "Hello World";
obj;

const arr = [
    1,2,3,4,5,6,7,8,9,10
];

arr[0];

xh  *= 2;
xh  /= 3;
xh  %= 4;
xh  += 3;
xh  -= 4;
xh >>= 5;
xh <<= 6;
xh  &= 7;
xh  |= 8;
xh  ^= 9;

var r = 0;

func mutate() {
    r;
    r = 200;
    local a = 300 + r;
    a;
    fact(3 + 2);
    return a;
}

const b = mutate();
b;
r;
b;
r;
fact(40);
```

## Architecture

### Compiler Pipeline

1. **Tokenizer** (`atom_tokenizer.go`): Converts source code into tokens
   - Supports Unicode strings with escape sequences (`\n`, `\t`, `\r`, `\uXXXX`)
   - Handles single-line (`//`) and multi-line (`/* */`) comments
   - Recognizes hexadecimal numbers (`0x1A`, `0xFF`)
   - Supports scientific notation (`1.5e10`, `2E-3`)
   - Multi-character operators (`==`, `!=`, `&&`, `||`, `+=`, `-=`, etc.)

2. **Parser** (`atom_parser.go`): Builds Abstract Syntax Tree (AST) from tokens
   - Recursive descent parser with operator precedence
   - Supports all binary operators with proper precedence
   - Handles function calls, array/object access, and member access
   - Parses control flow statements and variable declarations

3. **Compiler** (`atom_compile.go`): Compiles AST to bytecode
   - Generates stack-based bytecode instructions
   - Handles variable scoping and function compilation
   - Optimizes expression evaluation order

4. **Interpreter** (`atom_interpreter.go`): Executes bytecode
   - Stack-based virtual machine execution
   - Frame management for function calls
   - Memory management with garbage collection roots

### Runtime System

- **Stack-based VM**: Uses evaluation stack for expression evaluation
- **Frame Management**: Handles function calls and local scopes
- **Memory Management**: Basic garbage collection with GC roots
- **Type System**: Dynamic typing with runtime type checking

### Bytecode Operations

The VM supports the following opcodes:
- **Load Operations**: `OpLoadInt`, `OpLoadNum`, `OpLoadStr`, `OpLoadBool`, `OpLoadNull`, `OpLoadArray`, `OpLoadObject`, `OpLoadFunction`
- **Arithmetic**: `OpAdd`, `OpSub`, `OpMul`, `OpDiv`, `OpMod`
- **Bitwise**: `OpAnd`, `OpOr`, `OpXor`, `OpShl`, `OpShr`
- **Comparison**: `OpCmpLt`, `OpCmpLte`, `OpCmpGt`, `OpCmpGte`, `OpCmpEq`, `OpCmpNe`
- **Control Flow**: `OpJump`, `OpJumpIfFalseOrPop`, `OpJumpIfTrueOrPop`, `OpPopJumpIfFalse`, `OpPopJumpIfTrue`
- **Function Calls**: `OpCall`, `OpReturn`
- **Variable Access**: `OpLoadLocal`, `OpStoreLocal`, `OpLoadCapture`, `OpStoreCapture`, `OpStoreGlobal`
- **Object/Array Operations**: `OpIndex`, `OpSetIndex`
- **Stack Operations**: `OpDupTop`, `OpPopTop`, `OpNoOp`

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

- No object-oriented programming (no classes, inheritance, or `this` context)
- Limited array operations (no built-in methods like `push`, `pop`, `length`)
- No string manipulation functions
- Limited I/O (basic `print`/`println` only via `atom:std`)
- No built-in standard library beyond `atom:std`
- No closures (functions cannot capture variables from outer scope)
- No type annotations or static type checking

## Future Enhancements

- Loop constructs (`for`, `while`, `do-while`)
- Object-oriented features (classes, inheritance, `this` context)
- Module system and imports
- Standard library with common functions
- Better error messages and debugging support
- Performance optimizations
- REPL (Read-Eval-Print Loop)
- Closures and lexical scoping
- Type annotations and static type checking
- String manipulation functions
- I/O operations (file system, console)
- Array and object built-in methods
- Exception handling with try-catch
- Garbage collection improvements

## License

This project is licensed under the terms specified in the LICENSE file.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests to improve the language implementation.
