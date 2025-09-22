# Atom Programming Language

Atom is a custom programming language implemented in Go, featuring a complete compiler and interpreter with support for functions, variables, control flow, objects, arrays, classes, enums, and various data types. It uses a stack-based virtual machine for execution.

## Features

Atom is a modern, dynamically-typed programming language with a comprehensive feature set:

### Core Language Features
- **Dynamic Typing**: Supports integers (32-bit), floats (64-bit), strings, booleans, arrays, objects, functions, classes, enums, promises, and errors
- **Function Support**: First-class functions with parameters, return statements, and async/await support
- **Variable Declarations**: `var`, `const`, and `local` variable declarations with proper scoping
- **Control Flow**: `if-else`, `switch` statements and expressions, `while`, `do-while`, `break`, `continue`
- **Object-Oriented Programming**: Classes with inheritance (`extends`), methods, constructors, and instance management
- **Enums**: Enum declarations with custom values and pattern matching in switch expressions

### Advanced Features
- **Async/Await**: Full async/await support with promise-based execution and microtask scheduling
- **Error Handling**: Catch expressions with error propagation and comprehensive error reporting
- **Arithmetic Operations**: Full set of mathematical, bitwise, and logical operations with proper precedence
- **Assignment Operators**: Compound assignment operators (`+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=`)
- **For Loops**: For loop support with local variable scoping and increment operations
- **Closures**: Basic closure support with variable capture from outer scopes
- **Number Literals**: Support for hexadecimal (`0xFF`), binary (`0b1010`), and octal (`0o2`) number formats

### Data Structures
- **Array Support**: Array literals, indexing, element access, and freezing for immutability
- **Array Methods**: Built-in array methods including `push`, `select`, `where`, `any`, `all`, and `length`
- **Object Support**: Object literals with key-value pairs, property access, method calls, and freezing
- **Object Methods**: Built-in object methods including `keys` and `values` from `atom:object`
- **Self-Reference**: Support for circular references in arrays and objects

### Runtime Features
- **Stack-based VM**: Efficient bytecode execution with evaluation stack and frame management
- **Memory Management**: Basic garbage collection with GC roots and memory usage reporting
- **Imports & Stdlib**: Import from `atom:std` (e.g., `print`, `println`, `freeze`, `decompile`, `throw`, `readLine`) with colored output
- **Math Module**: Import from `atom:math` (e.g., `rand`, `sqrt`, `pow`) for mathematical operations
- **Object Module**: Import from `atom:object` (e.g., `keys`, `values`) for object manipulation
- **Error Reporting**: Comprehensive error messages with source location, context display, and debugging information
- **Module System**: Support for importing from `atom:std`, `atom:math`, `atom:object`, local modules, and relative imports

### Startup Experience
When you run Atom without arguments, it displays an impressive ASCII art banner:

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║    █████╗ ████████╗ ██████╗ ███╗   ███╗    ██████╗ ██████╗  ██████╗ ███████╗ ║
║   ██╔══██╗╚══██╔══╝██╔═══██╗████╗ ████║   ██╔════╝██╔═══██╗██╔═══██╗██╔════╝ ║
║   ███████║   ██║   ██║   ██║██╔████╔██║   ██║     ██║   ██║██║   ██║███████╗ ║
║   ██╔══██║   ██║   ██║   ██║██║╚██╔╝██║   ██║     ██║   ██║██║   ██║██╔════╝ ║
║   ██║  ██║   ██║   ╚██████╔╝██║ ╚═╝ ██║   ╚██████╗╚██████╔╝╚██████╔╝███████╗ ║
║   ╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝     ╚═╝    ╚═════╝ ╚═════╝  ╚═════╝ ╚══════╝ ║
║                                                                              ║
║                    A Custom Programming Language                             ║
║                    Implemented in Go                                         ║
║                                                                              ║
║  Features: Dynamic Typing • OOP • Functions • Arrays • Objects • Classes     ║
║  Author:   Philipp Andrew Redondo                                            ║
║  License:  MIT License                                                       ║
║  GitHub:   https://github.com/HolliShake/atomv3                              ║
║                                                                              ║
║  usage: atom <file.atom>                                                     ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Language Syntax

### Data Types

Atom supports the following primitive and complex types:
- **Integers**: `42`, `-17`, `0` (32-bit signed integers)
- **Floats**: `3.14`, `-2.5`, `0.0` (64-bit floating point numbers)
- **Number Literals**: Support for different number formats:
  - Decimal: `42`, `-17`, `0`
  - Hexadecimal: `0xFF`, `0x1A`, `0x0`
  - Binary: `0b1010`, `0b1111`, `0b0`
  - Octal: `0o2`, `0o777`, `0o0`
- **Strings**: `"hello"`, `"world"` (Unicode strings with escape sequences `\n`, `\t`, `\r`, `\uXXXX`)
- **Booleans**: `true`, `false`
- **Null**: `null`
- **Arrays**: `[1, 2, 3, 4, 5]` (dynamic arrays with indexing and freezing)
- **Objects**: `{key: value, name: "test"}` (key-value pairs with property access and freezing)
- **Functions**: First-class function objects with closure support
- **Classes**: User-defined types with methods, inheritance, and instance management
- **Enums**: Named constants with optional custom values and pattern matching
- **Promises**: Async operation results with state management (pending, fulfilled, rejected)
- **Errors**: Runtime error values for exception handling with detailed error messages

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

// Async function with await
async func fetchData(id) {
    println("Fetching data for ID:", id);
    return "Data-" + id;
}

// Async function calling other async functions
async func processData() {
    const data1 = await fetchData("001");
    const data2 = await fetchData("002");
    return [data1, data2];
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

// Number literals in different formats
var decimal = 42;       // Decimal
var hex = 0xFF;         // Hexadecimal (255)
var binary = 0b1010;    // Binary (10)
var octal = 0o2;        // Octal (2)

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
    
    func init(self, _name, _age) {
        self.name = _name;
        self.age = _age;
        println("Dog created: " + self.name);
    }
    
    func bark(self) {
        println("Woof! I'm " + self.name);
    }
}

// Creating instances with constructor parameters
var myDog = new Dog("Buddy", 3);
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

### Async/Await Support

```atom
// Basic async function
async func fetchUserData(userId) {
    println("Fetching user data for:", userId);
    return { id: userId, name: "User" + userId };
}

// Async function with multiple awaits
async func processUser(userId) {
    const userData = await fetchUserData(userId);
    const profile = await fetchUserProfile(userData.id);
    return { user: userData, profile: profile };
}

// Error handling in async functions
async func mayFail(shouldFail) {
    if (shouldFail) {
        return null; // Simulate failure
    }
    return "Success!";
}

// Async function with error handling
async func handleAsync() {
    const result = await mayFail(false);
    if (result == null) {
        println("Operation failed");
        return null;
    }
    return result;
}

// Calling async functions
processUser("123");
handleAsync();
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

// Error handling with array access
const arr = [1, 2, 3];
const value = arr[10] catch(e) {
    println("Array index error:", e);
    return -1;
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
import [ print, println, freeze, decompile, throw, readLine ] from "atom:std";

println("Hello, World!", 1, 2, 3);  // Output in yellow
print("Wheels on the bus!", std, "\n");  // Output in green

// Freezing objects and arrays
var obj = {name: "test", value: 42};
var arr = [1, 2, 3, 4, 5];

std.freeze(obj);
std.freeze(arr);

// These would cause errors:
// obj.name = "new name";  // Error: cannot set index on frozen object
// arr[0] = 999;          // Error: cannot set index on frozen array

// Function decompilation
func add(a, b) {
    return a + b;
}
println(std.decompile(add));  // Shows bytecode representation

// Error handling
const result = riskyOperation() catch(err) {
    std.throw(err);  // Throws error with stack trace
    return null;
};

// Interactive input
const name = std.readLine("Enter your name: ");
println("Hello, " + name + "!");

const Log = println; // aliasing
Log("Hello, World!");
```

### Standard Library Functions

Atom provides several built-in functions through multiple modules:

#### `atom:std` Module
- **`print(...)`**: Prints arguments in green color without newline
- **`println(...)`**: Prints arguments in yellow color with newline
- **`freeze(obj)`**: Makes objects and arrays immutable
- **`decompile(func)`**: Returns bytecode representation of a function
- **`throw(error)`**: Throws an error with full stack trace
- **`readLine(prompt)`**: Reads a line of input from stdin with optional prompt

#### `atom:math` Module
- **`rand(max)`**: Generates random numbers up to the specified maximum
- **`sqrt(num)`**: Calculates square root of a number
- **`pow(base, exponent)`**: Raises base to the power of exponent

#### `atom:object` Module
- **`keys(obj)`**: Returns array of object property keys
- **`values(obj)`**: Returns array of object property values
- **`freeze(obj)`**: Makes objects immutable (alternative to std.freeze)

### Module System

Atom supports a flexible module system with multiple import patterns:

```atom
// Import specific functions from standard library
import [ print, println, freeze ] from "atom:std";

// Import entire standard library
import "atom:std";

// Import math functions
import [ rand, sqrt, pow ] from "atom:math";

// Import object functions
import [ keys, values ] from "atom:object";

// Import local modules
import "gui";
import "hash";
import "./imports.atom";

// Using imported modules
std.println("Using std module");
math.rand(1000);  // Generate random number
math.sqrt(4);     // Calculate square root
math.pow(2, 5);   // Calculate power
object.keys({H: 200});  // Get object keys
gui.getName();  // Call function from gui module
hash.someFunction();  // Call function from hash module
```

The module system supports:
- **Standard Library**: `atom:std`, `atom:math`, and `atom:object` modules
- **Local Modules**: Files in the same directory or subdirectories
- **Relative Imports**: Using `./` for relative paths
- **Named Imports**: Import specific functions with destructuring syntax
- **Default Imports**: Import entire modules

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

### Loops: while, do-while, and for

```atom
// While loop
var i = 0;
while (1 && i < 10) {
    println(i);
    i += 1;
}

// Do-while loop
do {
    println(i);
    i += 1;
} while (i < 20);

// For loop with local variable scoping
for (local x = 0; x <= 10; x += 1) {
    println(">>", x);
}
```

## Example Programs

### Hello World
```atom
import [println] from "atom:std";
println("Hello, World!");
```

### Basic Arithmetic and Functions
```atom
import [println] from "atom:std";

func add(a, b) {
    return a + b;
}

func factorial(n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

println("5 + 3 =", add(5, 3));
println("Factorial of 5 =", factorial(5));
```

### Fibonacci Sequence
```atom
import [println] from "atom:std";

func fibonacci(n) {
    if (n <= 1) {
        return n;
    }
    return fibonacci(n - 1) + fibonacci(n - 2);
}

const n = 10;
println("Fibonacci sequence up to", n, ":", fibonacci(n));
```

### Object-Oriented Programming
```atom
import [println] from "atom:std";

class Animal {
    func init(self, name) {
        self.name = name;
    }
    
    func eat(self, food) {
        println(self.name + " is eating " + food);
    }
}

class Dog extends Animal {
    func init(self, name, age) {
        self.name = name;
        self.age = age;
    }
    
    func bark(self) {
        println("Woof! I'm " + self.name + " and I'm " + self.age + " years old");
    }
}

const myDog = new Dog("Buddy", 3);
myDog.eat("bone");
myDog.bark();
```

### Async/Await Programming
```atom
import [println] from "atom:std";

async func fetchData(id) {
    println("Fetching data for ID:", id);
    return "Data-" + id;
}

async func processData() {
    println("Starting data processing...");
    const data1 = await fetchData("001");
    const data2 = await fetchData("002");
    println("All data:", data1, data2);
    return [data1, data2];
}

processData();
```

### Error Handling with Catch Expressions
```atom
import [println] from "atom:std";

func riskyOperation() {
    return someArray[999]; // This might throw an error
}

const result = riskyOperation() catch(err) {
    println("Error occurred:", err);
    return null;
};

// String indexing with error handling
const g = "He😊lo"[2] catch(e) {
    println("String index error:", e);
    return "Hello!";
};

println("Result:", g);
```

### Switch Statements and Enums
```atom
import [println] from "atom:std";

enum Status {
    OK = 2,
    BAD = 3,
    PENDING = 1
}

var result = (Status.BAD) switch {
    case(Status.OK) => "Success"
    case(Status.BAD) => "Error"
    default => "Unknown"
};

println("Status result:", result);
println("Status values:", Status.OK, Status.BAD, Status.PENDING);
```

### Arrays and Objects with Built-in Methods
```atom
import [println, freeze] from "atom:std";
import [keys, values] from "atom:object";
import [rand, sqrt, pow] from "atom:math";

// Array operations with built-in methods
const numbers = [1, 2, 3, 4, 5];
numbers.push(6).push(7).push(8);
println("Array after push:", numbers);

// Array filtering and mapping
const filtered = numbers.where(function(n) {
    return n > 5;
});
println("Filtered array:", filtered);

const mapped = numbers.select(function(n, index) {
    return n * 2;
});
println("Mapped array:", mapped);

// Array testing methods
println("All > 0:", numbers.all(function(x) { return x > 0; }));
println("Any > 10:", numbers.any(function(x) { return x > 10; }));
println("Array length:", numbers.length());

// Object operations with built-in methods
const person = {
    name: "Alice",
    age: 30,
    city: "New York",
    greet: function() {
        return "Hello, I'm " + this.name;
    }
};

println("Object keys:", keys(person));
println("Object values:", values(person));
println("Greeting:", person.greet());

// Math operations
println("Random number:", rand(100));
println("Square root of 16:", sqrt(16));
println("2 to the power of 8:", pow(2, 8));

// Freezing objects and arrays
freeze(person);
freeze(numbers);
// person.name = "Bob"; // This would cause an error
```

### Complex Control Flow
```atom
import [println] from "atom:std";

var counter = 0;

// While loop with break and continue
while (counter < 10) {
    if (counter == 5) {
        counter += 1;
        continue;
    }
    if (counter == 8) {
        break;
    }
    println("Counter:", counter);
    counter += 1;
}

// Do-while loop
do {
    println("Do-while iteration:", counter);
    counter += 1;
} while (counter < 12);

// Conditional expressions
const value = if (counter > 10) 
    100 
else 
    50;

println("Conditional value:", value);
```

### Advanced Async Patterns
```atom
import [println] from "atom:std";

async func calculate(a, b) {
    println("Calculating:", a, "+", b);
    return a + b;
}

async func processWithLoops() {
    local results = [null, null, null];
    local index = 0;
    
    while (index < 3) {
        const data = await calculate(index, index * 2);
        results[index] = data;
        println("Loop iteration:", index, "result:", data);
        index = index + 1;
    }
    
    return results;
}

processWithLoops();
```

### Comprehensive Test Example
```atom
import [print, println] from "atom:std";

// This example demonstrates most language features
var x = 1;

func add(a, b) {
    return a + b;
}

// Arithmetic expressions
2 + 23 / 23 - 2;
3 + 2 + add(12, 2);

// Control flow
if (0 || 0) {
    2 + 2;
} else {
    3 + 3;
    local x = 69420;
    local y = x + 1;
    x;
    y;
}

// Recursive function
func fact(n) {
    if (n == 0) {
        return 1;
    }
    return n * fact(n - 1);
}

const n = 3 + 2;
fact(n);

// Arrays and objects
println([1, 2, 3, 4, 5, 6, fact(n), "Nah!!"]);

var obj = {
    getName: "Hello",
    add: add,
    val: 0
};

obj.val += 2;
obj["val"] += 2;
println(obj);

// Assignment operators
var xh = 12;
xh *= 2;
xh /= 3;
xh %= 4;
xh += 3;
xh -= 4;
xh >>= 5;
xh <<= 6;
xh &= 7;
xh |= 8;
xh ^= 9;

// Function with side effects
var r = 1;
func mutate() {
    println(r);
    r += 200;
    local a = 300 + r;
    return a;
}

const b = mutate();
println("Final r:", r);

// Switch expressions
const xx = (5) switch {
    case(1, 2) => 200
    case(3, 4) => 300
    default => 400
};

println("Switch result:", xx);

// Conditional expressions
const cond = if (3 < 2)
    2 + 2
else if (3 < 2)
    3 + 3
else
    4 + 4;

println("Conditional result:", cond);

println("Done");
```

## Test Files and Examples

The `test/` directory contains comprehensive test files that demonstrate various language features and capabilities:

### Core Test Files

- **`test.atom`** (273 lines) - The main comprehensive test suite that demonstrates:
  - Basic arithmetic operations and operator precedence
  - Function declarations and recursive calls
  - Control flow with if-else statements
  - Variable declarations (`var`, `const`, `local`)
  - Object creation and manipulation
  - Array operations and indexing
  - Assignment operators (`+=`, `-=`, `*=`, etc.)
  - Switch statements and expressions
  - Conditional expressions
  - While and do-while loops with break/continue
  - Error handling with catch expressions
  - Enum declarations and usage
  - Class definitions and inheritance
  - Import statements from `atom:std` and `atom:object`
  - Function decompilation with `std.decompile()`
  - Unicode string handling with emoji characters
  - Method binding and chaining

- **`arithmetic.atom`** (26 lines) - Tests basic arithmetic operations:
  - Function calls with parameters and return values
  - Addition operations with formatted output
  - Class instantiation with constructors
  - Basic object creation and property access

- **`array.atom`** (12 lines) - Demonstrates array features:
  - Array literals with multiple elements
  - Array self-reference (circular reference)
  - Array indexing and modification
  - Array printing and display

- **`async.atom`** (368 lines) - Comprehensive async/await testing:
  - Basic async function chains and sequential execution
  - Multiple concurrent async calls with different return types
  - Error handling in async functions with null returns
  - Sequential async processing with message building
  - Different return types from async functions (numbers, arrays, objects)
  - Nested async calls (3+ levels deep with level1/level2/level3)
  - Async functions with parameters and calculations
  - Async operations within while loops with counter management
  - Async operations within do-while loops with accumulation
  - Nested while loops with async calls and result collection
  - Conditional logic in async loops with success/failure patterns
  - Accumulation patterns with async (sum, product, string building)
  - String building with async operations and delimiter handling
  - Array building using async operations with mixed data types
  - Complex nested async structures with multi-level processing

- **`class.atom`** (137 lines) - Advanced object-oriented programming features:
  - Base class definitions with methods and constructors
  - Class inheritance using `extends` with method overriding
  - Constructor methods with parameters and global variable access
  - Async methods in classes with await support
  - Error handling with division by zero and catch expressions
  - Class instantiation and method calls with chaining
  - Module imports (`gui`, `hash`, `imports.atom`, `atom:math`)
  - Function decompilation examples with bytecode display
  - For loops with local variable scoping
  - Array methods (push, select, where, any, all, length)
  - Object methods (keys, values) from `atom:object`
  - Math functions (rand, sqrt, pow) from `atom:math`
  - Closure support with variable capture
  - Hexadecimal, binary, and octal number literals

- **`function.atom`** (20 lines) - Function capabilities:
  - Recursive function definitions with base cases
  - Factorial calculation with proper recursion
  - Fibonacci sequence calculation
  - Function calls with return values and std module usage

- **`object.atom`** (24 lines) - Object manipulation:
  - Object literals with properties and method definitions
  - Object freezing functionality with immutability
  - Self-referencing objects with circular references
  - Property access and modification with bracket notation
  - Object comparison and equality testing

- **`imports.atom`** (17 lines) - Module system testing:
  - Global variable access and mutation
  - Recursive function with error handling and catch expressions
  - Error throwing and catching with std.throw()
  - Module import functionality and std module usage

- **`scope.atom`** (42 lines) - Variable scoping and block scope:
  - Global variable mutation across function boundaries
  - Block scope with curly braces and local variable isolation
  - Function scope and variable access patterns
  - Class instantiation and method chaining with return values
  - Recursive function calls with proper scoping

### Library Files

- **`app/lib/hash.atom`** - Simple library module demonstrating module exports
- **`app/lib/gui/index.atom`** - GUI library module with exported functions

### Comparison Files

- **`async.js`** (440 lines) - JavaScript equivalent of async.atom for comparison
- **`async.py`** (320 lines) - Python equivalent of async.atom for comparison
- **`class.py`** (25 lines) - Python equivalent of class.atom for comparison
- **`scope.js`** (19 lines) - JavaScript equivalent of scope.atom for comparison

These comparison files demonstrate how Atom's syntax compares to established languages like JavaScript and Python, showing the language's modern programming capabilities.

### Running Tests

To run any test file:

```bash
# Run the comprehensive test suite
./atom test/test.atom

# Run specific feature tests
./atom test/async.atom
./atom test/class.atom
./atom test/arithmetic.atom
```

Each test file is designed to be self-contained and demonstrates specific language features, making them excellent examples for learning the language syntax and capabilities.

## Architecture

### Compiler Pipeline

1. **Tokenizer** (`tokenizer.go`): Converts source code into tokens
   - Supports Unicode strings with escape sequences (`\n`, `\t`, `\r`, `\uXXXX`)
   - Handles single-line (`//`) and multi-line (`/* */`) comments
   - Recognizes hexadecimal numbers (`0x1A`, `0xFF`)
   - Supports scientific notation (`1.5e10`, `2E-3`)
   - Multi-character operators (`==`, `!=`, `&&`, `||`, `+=`, `-=`, etc.)
   - Identifies keywords, identifiers, and symbols
   - Position tracking for accurate error reporting

2. **Parser** (`parser.go`): Builds Abstract Syntax Tree (AST) from tokens
   - Recursive descent parser with operator precedence
   - Supports all binary operators with proper precedence
   - Handles function calls, array/object access, and member access
   - Parses control flow statements and variable declarations
   - Supports classes, enums, switch statements, catch expressions, and async/await
   - Comprehensive error handling with context display

3. **Compiler** (`compile.go`): Compiles AST to bytecode
   - Generates stack-based bytecode instructions
   - Handles variable scoping and function compilation
   - Optimizes expression evaluation order
   - Manages symbol tables and scope resolution
   - Supports async function compilation with promise handling

4. **Interpreter** (`interpreter.go`): Executes bytecode
   - Stack-based virtual machine execution
   - Frame management for function calls
   - Memory management with garbage collection roots
   - Runtime error handling and reporting
   - Integration with async scheduler for promise-based execution

### Runtime System

- **Stack-based VM**: Uses evaluation stack for expression evaluation with efficient memory management
- **Frame Management**: Handles function calls, local scopes, and async execution contexts
- **Async Scheduler**: Microtask-based scheduling system for handling async/await operations
- **Memory Management**: Basic garbage collection with GC roots and memory usage reporting
- **Type System**: Dynamic typing with runtime type checking and comprehensive error messages
- **Promise System**: Full promise implementation with state management (pending, fulfilled, rejected)

### Async Scheduler and Promise System

Atom implements a sophisticated async/await system with the following components:

- **Microtask Scheduler**: Handles async operations using a microtask queue similar to JavaScript's event loop
- **Promise States**: Promises can be in `pending`, `fulfilled`, or `rejected` states
- **Execution States**: Async functions have execution states (`ExecIdle`, `ExecAwaiting`, `ExecRunning`, `ExecCompleted`)
- **Frame Management**: Async functions maintain their own execution frames with promise references
- **Suspension/Resumption**: The scheduler can suspend and resume async function execution
- **Error Propagation**: Async errors are properly propagated through the promise chain

The async system allows for:
- Non-blocking async function execution
- Proper handling of multiple concurrent async operations
- Promise-based error handling
- Integration with the main execution loop

### Bytecode Operations

The VM supports the following opcodes:
- **Load Operations**: `OpLoadInt`, `OpLoadNum`, `OpLoadStr`, `OpLoadBool`, `OpLoadNull`, `OpLoadArray`, `OpLoadObject`, `OpLoadName`, `OpLoadCapture`, `OpLoadModule`, `OpLoadFunction`
- **Arithmetic**: `OpAdd`, `OpSub`, `OpMul`, `OpDiv`, `OpMod`
- **Bitwise**: `OpAnd`, `OpOr`, `OpXor`, `OpShl`, `OpShr`
- **Comparison**: `OpCmpLt`, `OpCmpLte`, `OpCmpGt`, `OpCmpGte`, `OpCmpEq`, `OpCmpNe`
- **Control Flow**: `OpJump`, `OpAbsoluteJump`, `OpJumpIfFalseOrPop`, `OpJumpIfTrueOrPop`, `OpPopJumpIfFalse`, `OpPopJumpIfTrue`, `OpPeekJumpIfEqual`, `OpPopJumpIfNotError`
- **Function Calls**: `OpCall`, `OpCallConstructor`, `OpReturn`, `OpAwait`
- **Variable Access**: `OpStoreModule`, `OpStoreCapture`, `OpStoreLocal`
- **Object/Array Operations**: `OpIndex`, `OpSetIndex`, `OpPluckAttribute`
- **Class/Enum Operations**: `OpMakeClass`, `OpExtendClass`, `OpMakeEnum`
- **Stack Operations**: `OpDupTop`, `OpPopTop`, `OpNoOp`
- **Unary Operations**: `OpNot`, `OpNeg`, `OpPos`, `OpTypeof`
- **Module Operations**: `OpMakeModule`

## Building and Running

### Prerequisites
- Go 1.24.5 or later

### Build
```bash
# Build the compiler/interpreter
cd app

# Using the build script (recommended)
./build.sh        # Linux/macOS
# or
build.bat         # Windows

# Manual build
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

### Debugging and Error Reporting

Atom provides comprehensive error reporting with:
- **Source Location**: Shows exact file, line, and column where errors occur
- **Context Display**: Shows surrounding code with error highlighting
- **Memory Usage**: Reports memory allocation statistics after execution
- **Call Stack**: Shows function call hierarchy in error messages

Example error output:
```
DEBUG(parseExpression)::Error in [test.atom:15:8] unexpected token
  12 | var x = 1;
  13 | var y = 2;
  14 | var z = 3;
  15 | var w = ;  // Missing value
     |        ^
  16 | println(x);
```

## Project Structure

```
atom/
├── app/                   # Main application and compiler
│   ├── main.go            # Entry point with startup banner
│   ├── tokenizer.go       # Lexical analysis with Unicode support
│   ├── parser.go          # Syntax analysis and AST generation
│   ├── compile.go         # Code generation to bytecode
│   ├── ast.go             # AST node definitions and types
│   ├── error.go           # Error handling and reporting
│   ├── keyword.go         # Language keywords and constants
│   ├── token.go           # Token definitions and types
│   ├── position.go        # Source position tracking
│   ├── scope.go           # Variable scoping and symbol tables
│   ├── symbol.go          # Symbol table management
│   ├── atom.ico           # Application icon
│   ├── atom.linux         # Linux binary executable
│   ├── build.sh           # Build script (Linux/macOS)
│   ├── build.bat          # Build script (Windows)
│   ├── go.mod             # App module dependencies
│   ├── go.sum             # App module checksums
│   └── lib/               # Library modules
│       ├── gui/
│       │   └── index.atom # GUI library module
│       └── hash.atom      # Hash library module
├── runtime/               # Runtime system and VM
│   ├── interpreter.go     # VM implementation and execution
│   ├── operation.go       # Bytecode operation implementations
│   ├── value.go           # Value system and type definitions
│   ├── opcode.go          # Bytecode opcode definitions
│   ├── stack.go           # Evaluation stack implementation
│   ├── state.go           # Runtime state management
│   ├── code.go            # Code generation utilities
│   ├── object.go          # Object system and properties
│   ├── array.go           # Array implementation
│   ├── class.go           # Class definitions and inheritance
│   ├── class_instance.go  # Class instance handling
│   ├── checker.go         # Type checking utilities
│   ├── util.go            # Utility functions
│   ├── builtin.go         # Built-in functions
│   ├── builtin_std.go     # Standard library functions (print, println, freeze, decompile, throw, readLine)
│   ├── builtin_object.go  # Object-related built-in functions
│   ├── frame.go           # Call frame management
│   ├── scheduler.go       # Async scheduler for promises
│   ├── promise.go         # Promise implementation
│   ├── decompile.go       # Bytecode decompilation utilities
│   ├── cell.go            # Cell implementation for closures
│   ├── debug.go           # Debug utilities
│   ├── go.mod             # Runtime module dependencies
│   └── go.sum             # Runtime module checksums
├── test/                  # Test files and examples
│   ├── test.atom          # Comprehensive test suite (273 lines)
│   ├── arithmetic.atom    # Basic arithmetic and function calls (26 lines)
│   ├── array.atom         # Array operations and self-reference (12 lines)
│   ├── async.atom         # Advanced async/await patterns (368 lines)
│   ├── class.atom         # Class inheritance and async methods (66 lines)
│   ├── function.atom      # Recursive functions and factorial (20 lines)
│   ├── object.atom        # Object freezing and self-reference (24 lines)
│   ├── imports.atom       # Module import testing (17 lines)
│   ├── scope.atom         # Variable scoping and block scope (42 lines)
│   ├── async.js           # JavaScript async comparison (440 lines)
│   ├── async.py           # Python async comparison (320 lines)
│   ├── class.py           # Python class comparison (25 lines)
│   └── scope.js           # JavaScript scope comparison (19 lines)
├── go.work                # Go workspace configuration
├── go.work.sum            # Go workspace checksums
├── readme.MD              # This documentation
└── LICENSE                # MIT License
```

## Language Limitations

Based on the test files and implementation analysis, the following limitations are currently present:

### Data Structure Limitations
- **Array Methods**: While some array methods exist (`push`, `select`, `where`, `any`, `all`, `length`), the full set of common array operations is incomplete (no `pop`, `shift`, `unshift`, `slice`, `splice`, `indexOf`, etc.)
- **String Operations**: No string manipulation functions (`substring`, `split`, `join`, `replace`, `trim`, `toUpperCase`, `toLowerCase`, etc.)
- **Advanced Data Structures**: No built-in data structures beyond arrays and objects (no sets, maps, queues, stacks, trees, etc.)

### Control Flow Limitations
- **For Loops**: While `for` loops are supported in `class.atom`, they may have limited functionality compared to traditional for loops
- **Iterator Support**: No `foreach` or iterator-based loops for arrays/objects
- **Exception Handling**: Limited exception handling beyond catch expressions (no try-catch blocks or comprehensive error recovery)

### I/O and System Limitations
- **Basic I/O**: Limited I/O operations (basic `print`/`println`/`freeze`/`decompile`/`throw`/`readLine` via `atom:std`)
- **File Operations**: No file I/O operations (read/write files, directory operations, file system access)
- **Network Operations**: No network operations or HTTP client functionality
- **Standard Library**: Limited standard library (only `atom:std`, `atom:object`, and `atom:math` modules)

### Language Feature Limitations
- **Closures**: Limited closure support (some variable capture exists but may not be fully implemented)
- **Type System**: No type annotations or static type checking
- **Module System**: Basic module system (imports from `atom:std`, local modules, and relative imports)
- **Generics**: No generics or template programming
- **Pattern Matching**: Limited pattern matching beyond switch expressions
- **Operator Overloading**: No operator overloading for custom types

### Development and Debugging Limitations
- **Debugging Tools**: Limited debugging tools and profiling support
- **REPL**: No REPL (Read-Eval-Print Loop) for interactive development
- **Testing Framework**: No built-in testing framework
- **Code Quality**: No code formatting or linting tools
- **Error Reporting**: Basic error messages and stack traces (improved with `std.throw`)

### Performance Limitations
- **Compilation**: No JIT compilation or performance optimizations
- **Concurrency**: No parallel execution or threading support
- **Memory Management**: Basic garbage collection without advanced memory management
- **Code Organization**: No code splitting or lazy loading capabilities

### Runtime Limitations
- **Native Binding**: No native function binding beyond built-in functions
- **Memory Management**: Limited memory management (basic GC with roots)
- **Reflection**: Limited reflection or introspection capabilities beyond `std.decompile`
- **Serialization**: No serialization/deserialization support (JSON, binary formats, etc.)

### Advanced Features Missing
- **Regular Expressions**: No regex support for string pattern matching
- **Date/Time**: No built-in date and time handling
- **Random Numbers**: Basic random number generation exists but may be limited
- **Mathematical Functions**: Limited mathematical functions beyond basic operations
- **Unicode Support**: Basic Unicode support exists but may be incomplete for complex operations

## Future Enhancements

### High Priority
- **String Operations**: String manipulation functions (`substring`, `split`, `join`, `replace`, `trim`, `toUpperCase`, `toLowerCase`, etc.)
- **Array Methods**: Complete array method set (`pop`, `shift`, `unshift`, `slice`, `splice`, `indexOf`, `lastIndexOf`, etc.)
- **I/O Operations**: File I/O operations (read/write files, directory operations, file system access)
- **REPL**: Interactive Read-Eval-Print Loop for development
- **Error Handling**: Enhanced exception handling with try-catch blocks and comprehensive error recovery

### Medium Priority
- **Module System**: Enhanced module system with better import/export and dependency management
- **Closures**: Full lexical scoping and closure support for functions
- **Debugging**: Better error messages, debugging support, and profiling tools
- **Testing**: Built-in testing framework and test runners
- **Advanced Data Structures**: Sets, maps, queues, stacks, and other data structures
- **Regular Expressions**: Regex support for string pattern matching

### Low Priority
- **Performance**: Performance optimizations and JIT compilation
- **Type System**: Type annotations and static type checking
- **Memory Management**: Improved garbage collection and memory optimization
- **Pattern Matching**: Enhanced pattern matching capabilities beyond switch expressions
- **Generics**: Generic programming support for reusable code
- **Concurrency**: Enhanced concurrency primitives and parallel execution
- **Native Binding**: Support for calling native functions and libraries
- **Serialization**: JSON and other serialization format support
- **Date/Time**: Built-in date and time handling
- **Network Operations**: HTTP client functionality and network operations

## Author and License

**Author**: Philipp Andrew Redondo  
**License**: MIT License  
**GitHub**: https://github.com/HolliShake/atomv3

This project is licensed under the terms specified in the LICENSE file.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests to improve the language implementation.

## Acknowledgments

Atom is a custom programming language implementation showcasing modern language design principles, including:
- Dynamic typing with runtime type checking
- Object-oriented programming with inheritance
- Functional programming with first-class functions
- Asynchronous programming with async/await
- Stack-based virtual machine architecture
- Comprehensive error handling and reporting

The implementation demonstrates the complexity and beauty of programming language design and implementation.

## Appendix: Architecture and Tests (Concise)

- **Compiler (app/)**:
  - `tokenizer.go` → tokens with Unicode, multi-char ops, number bases.
  - `parser.go` → AST for expressions/statements, classes, enums, async, catch.
  - `compile.go` → bytecode emission; scopes, captures, functions, modules.
- **Runtime (runtime/)**:
  - `opcode.go` → opcodes; `interpreter.go` → VM loop; `operation.go` → ops.
  - Async via `scheduler.go`/`promise.go`; objects/arrays/classes in respective files.
- **Execution flow**: source → tokens → AST → bytecode → interpreter executes.

- **Built-in modules**: `atom:std`, `atom:object`, `atom:math`. Import via:
  ```atom
  import [println] from "atom:std";
  ```

### Running tests quickly

```bash
# From repo root (binary in app/ if you built there)
./app/atom.linux test/test.atom | cat
./app/atom.linux test/async.atom | cat

# Or build and run
cd app && ./build.sh && ./atom ../test/test.atom | cat
# Direct with Go (from app/)
go run *.go ../test/test.atom | cat
```

Notes:
- Go version is defined in `app/go.mod` as 1.24.5; ensure your toolchain supports it.
- Library imports (non-`atom:`) resolve relative to executable `lib/` (see `app/lib/`).
- Tests are example scripts; there is no separate test runner.
