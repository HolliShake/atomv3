# Atom Programming Language

Atom is a custom programming language implemented in Go, featuring a complete compiler and interpreter with support for functions, variables, control flow, objects, arrays, classes, enums, and various data types. It uses a stack-based virtual machine for execution.

## Features

- **Dynamic Typing**: Supports integers, floats, strings, booleans, arrays, objects, functions, classes, and enums
- **Function Support**: First-class functions with parameters and return statements
- **Variable Declarations**: `var`, `const`, and `local` variable declarations with proper scoping
- **Control Flow**: `if-else`, `switch` statements and expressions, `while`, `do-while`, `break`, `continue`
- **Object-Oriented Programming**: Classes with inheritance (`extends`), methods, and constructors
- **Enums**: Enum declarations with custom values and pattern matching
- **Arithmetic Operations**: Full set of mathematical, bitwise, and logical operations
- **Assignment Operators**: Compound assignment operators (`+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=`)
- **Array Support**: Array literals, indexing, element access, and freezing
- **Object Support**: Object literals with key-value pairs, property access, method calls, and freezing
- **Error Handling**: Try-catch expressions with error propagation
- **Imports & Stdlib**: Import from `atom:std` (e.g., `print`, `println`, `freeze`)
- **Stack-based VM**: Efficient bytecode execution with evaluation stack
- **Memory Management**: Basic garbage collection with GC roots

## Language Syntax

### Data Types

Atom supports the following primitive types:
- **Integers**: `42`, `-17`, `0` (32-bit signed integers)
- **Floats**: `3.14`, `-2.5`, `0.0` (64-bit floating point numbers)
- **Strings**: `"hello"`, `"world"` (Unicode strings with escape sequences)
- **Booleans**: `true`, `false`
- **Null**: `null`
- **Arrays**: `[1, 2, 3, 4, 5]` (dynamic arrays with indexing and freezing)
- **Objects**: `{key: value, name: "test"}` (key-value pairs with property access and freezing)
- **Functions**: First-class function objects
- **Classes**: User-defined types with methods and inheritance
- **Enums**: Named constants with optional custom values

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

// While loops
var i = 0;
while (i < 10) {
    println(i);
    i += 1;
}

// Do-while loops
do {
    println(i);
    i += 1;
} while (i < 20);

// Break and continue
while (i < 20) {
    if (i == 15) {
        break;
    }
    if (i % 2 == 0) {
        continue;
    }
    println(i);
    i += 1;
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

// Freezing objects (making them immutable)
std.freeze(person);
// person.name = "Charlie"; // This would cause an error
```

### Classes

```atom
// Base class
class Animal {
    func eat(self) {
        println("Animal is eating");
    }
}

// Derived class with inheritance
class Dog extends Animal {
    local name = "Buddy";
    
    func init(self) {
        println("Dog created");
    }
    
    func bark(self) {
        println("Woof! I'm " + self.name);
    }
}

// Creating instances
var myDog = new Dog();
myDog.eat();    // Inherited method
myDog.bark();   // Own method
```

### Enums

```atom
// Enum declaration with custom values
enum Status {
    OK  = 2,
    BAD = 3,
    PENDING = 1
}

// Using enums
var result = (Status.BAD) switch {
    case(Status.OK)  => "Success"
    case(Status.BAD) => "Error"
    default => "Unknown"
};

println(result); // "Error"
println(Status.OK, Status.BAD); // 2 3
```

### Switch Statements and Expressions

```atom
// Switch as a statement
switch (2 + 2 + 3) {
    case(2, 4): {
        println("Hola");
    }
    default: {
        println("Mundo");
    }
}

// Switch as an expression
const value = (5) switch {
    case(1, 2) => 200
    case(3, 4) => 300
    default    => 400
};

println(">>", value); // 400
```

### Error Handling with Catch Expressions

```atom
// Try-catch expression
const g = "He😊lo"[2] catch(e) {
    println(">>", e);
    return "Hello!";
};

println(g); // "😊" or "Hello!" if error

// Error handling in function calls
func riskyOperation() {
    // This might throw an error
    return someArray[999];
}

const result = riskyOperation() catch(err) {
    println("Error occurred:", err);
    return null;
};
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

### Imports and Standard Library

```atom
// Import selected symbols from the standard library
import [ print, println, freeze ] from "atom:std";

println("Hello, World!", 1, 2, 3);
print("Wheels on the bus!", std, "\n");

// Freezing objects and arrays
var obj = {name: "test", value: 42};
var arr = [1, 2, 3, 4, 5];

std.freeze(obj);
std.freeze(arr);

// These would cause errors:
// obj.name = "new name";  // Error: cannot set index on frozen object
// arr[0] = 999;          // Error: cannot set index on frozen array

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
import [print, println] from "atom:std";

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

println([1, 2, 3, 4, 5, 6, fact(n), "Nah!!"]);

var obj = {
    getName: "Hello",
    add: add,
    val: 0
};

obj;
obj.getName;

obj.val += 2;
obj["val"] += 2;

println(obj);

var xh = 12;

xh *= 2;
xh = 100;

'2';;

std.freeze(obj);

println("error", obj.getName = "Hello World");
obj;

const arr = [
    1,2,3,4,5,6,7,8,9,10
];

std.freeze(arr);
println("error arr freeze", arr[0] = 2);

println("error", arr[232323] = 2);
println("error", 2[23]);

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

var r = 1;

func mutate() {
    r;
    println(r);
    r += 200;
    local a = 300 + r;
    a;
    fact(3 + 2);
    return a;
}

const b = mutate();
b;
println(r);
b;
r;

fact(40);

println("Hello, World!", 1, 2, 3);

print("Wheels on the bus!", std, "\n");

const mod = [];

const Log = println;

Log("Hello, World!");

println(!1);

const xx = (5) switch {
    case(1, 2) => 200
    case(3, 4) => 300
    default    => 400
};

println(">>", xx);

const cond = 
    if (3 < 2)
        2 + 2
    else if (3 < 2)
        3 + 3
    else
        4 + 4
    ;

println(cond);

var ll = 0;

while (1 && ll < 10) {
    println(ll);
    ll += 1;
}

do {
    if (ll == 18) {
        ll += 1;
        continue;
    } else if (ll == 19) {
        ll += 1;
        break;
    }
    println(ll);
    ll += 1;
} while (ll < 20)

println(ll);

println("Done");

func $test_getName() {
    return "Hello, World!";
}

const g = "He😊lo"[2] catch(e) {
    println(">>", e);
    return "Hello!";
};

println(g);

switch (2 + 2 + 3) {
    case(2, 4): {
        println("Hola");
    }
    default: {
        println("Mundo");
    }
}

println(true && false, null);

enum Status {
    OK  = 2,
    BAD = 3
}

var result = (Status.BAD) switch {
    case(Status.OK)  => "Yes"
    case(Status.BAD) => "No"
    default => "Unknown"
};

println(result);

println(">>", Status, Status.OK, Status.BAD);

class Animal {
    func eat(self) {

    }
}

class Dog extends Animal {
    local x = 2;

    func init(self) {

    }
}
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
- **Load Operations**: `OpLoadInt`, `OpLoadNum`, `OpLoadStr`, `OpLoadBool`, `OpLoadNull`, `OpLoadArray`, `OpLoadObject`, `OpLoadFunction`, `OpLoadModule0`
- **Arithmetic**: `OpAdd`, `OpSub`, `OpMul`, `OpDiv`, `OpMod`
- **Bitwise**: `OpAnd`, `OpOr`, `OpXor`, `OpShl`, `OpShr`
- **Comparison**: `OpCmpLt`, `OpCmpLte`, `OpCmpGt`, `OpCmpGte`, `OpCmpEq`, `OpCmpNe`
- **Control Flow**: `OpJump`, `OpJumpIfFalseOrPop`, `OpJumpIfTrueOrPop`, `OpPopJumpIfFalse`, `OpPopJumpIfTrue`, `OpPeekJumpIfEqual`, `OpPopJumpIfNotError`
- **Function Calls**: `OpCall`, `OpReturn`
- **Variable Access**: `OpLoadLocal`, `OpStoreLocal`, `OpStoreGlobal`
- **Object/Array Operations**: `OpIndex`, `OpSetIndex`, `OpPluckAttribute`
- **Class/Enum Operations**: `OpMakeClass`, `OpExtendClass`, `OpMakeEnum`
- **Stack Operations**: `OpDupTop`, `OpPopTop`, `OpNoOp`, `OpRot2`
- **Unary Operations**: `OpNot`, `OpNeg`, `OpPos`

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

- Limited array operations (no built-in methods like `push`, `pop`, `length`)
- No string manipulation functions
- Limited I/O (basic `print`/`println`/`freeze` only via `atom:std`)
- No built-in standard library beyond `atom:std`
- No closures (functions cannot capture variables from outer scope)
- No type annotations or static type checking
- No `for` loops (only `while` and `do-while`)
- No exception handling beyond catch expressions
- No module system beyond basic imports

## Future Enhancements

- `for` loops (currently only `while` and `do-while` are supported)
- Enhanced module system with better import/export
- Expanded standard library with common functions
- String manipulation functions (`substring`, `split`, `join`, etc.)
- Array and object built-in methods (`push`, `pop`, `length`, `keys`, `values`, etc.)
- File I/O operations (read/write files)
- Better error messages and debugging support
- Performance optimizations
- REPL (Read-Eval-Print Loop)
- Closures and lexical scoping
- Type annotations and static type checking
- Enhanced exception handling beyond catch expressions
- Garbage collection improvements
- Pattern matching enhancements
- Generic programming support

## License

This project is licensed under the terms specified in the LICENSE file.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests to improve the language implementation.
