# Atom Programming Language

Atom is a custom programming language implemented in Go, featuring a complete compiler and interpreter with support for functions, variables, control flow, objects, arrays, classes, async/await, and various data types. It uses a stack-based virtual machine for execution.

## Features

Atom is a modern, dynamically-typed programming language with a comprehensive feature set:

- **Dynamic Typing**: Variables can hold any type of value
- **Object-Oriented Programming**: Classes with methods and properties
- **Functional Programming**: First-class functions, closures, and higher-order functions
- **Asynchronous Programming**: Built-in async/await support
- **Collections**: Arrays and objects with rich APIs
- **Control Flow**: For loops, while loops, do-while loops, conditionals
- **Error Handling**: catch expression and error propagation
- **Modules**: Import system with standard library
- **Memory Management**: Automatic garbage collection

## Language Syntax and Semantics

### Variables and Constants

```atom
// Global variables
var globalVar = 42;
var globalString = "Hello, World!";

// Local variables (function-scoped)
local localVar = 100;
local localArray = [1, 2, 3];

// Constants (immutable)
const PI = 3.14159;
const CONFIG = { debug: true, version: "1.0" };
```

### Data Types

Atom supports several built-in data types:

- **Numbers**: Integers and floating-point numbers (`42`, `3.14`, `-10.5`)
- **Strings**: Text literals (`"Hello"`, `'World'`)
- **Booleans**: `true` and `false`
- **Null**: `null` for empty values
- **Arrays**: Ordered collections (`[1, 2, 3]`, `["a", "b", "c"]`)
- **Objects**: Key-value pairs (`{name: "John", age: 30}`)
- **Functions**: First-class functions

### Arithmetic Operations

```atom
// Basic arithmetic
local sum = 10 + 5;        // 15
local diff = 10 - 3;       // 7
local product = 4 * 6;     // 24
local quotient = 15 / 3;   // 5
local remainder = 10 % 3;  // 1

// Unary operations
local negative = -5;       // -5
local positive = +10;      // 10

// Complex expressions with operator precedence
local result = 2 + 3 * 4;     // 14 (multiplication first)
local grouped = (2 + 3) * 4;  // 20 (parentheses override)
```

### Control Flow

#### Conditional Statements

```atom
// If-else statements
if (condition) {
    // code block
} else if (anotherCondition) {
    // alternative code block
} else {
    // default code block
}

// Ternary-like operations
local result = condition ? value1 : value2;
```

#### Loops

```atom
// For loops
for (local i = 0; i < 10; i += 1) {
    println("Iteration:", i);
}

// While loops
local counter = 0;
while (counter < 5) {
    println("Counter:", counter);
    counter += 1;
}

// Do-while loops
local value = 0;
do {
    println("Value:", value);
    value += 1;
} while (value < 3);

// Loop control
for (local i = 0; i < 10; i += 1) {
    if (i == 5) break;        // Exit loop
    if (i % 2 == 0) continue; // Skip iteration
    println("Odd number:", i);
}
```

### Functions

#### Function Declaration and Invocation

```atom
// Basic function
func greet(name) {
    return "Hello, " + name + "!";
}

// Function with multiple parameters
func add(a, b) {
    return a + b;
}

// Function with no parameters
func getCurrentTime() {
    return "2024-01-01";
}

// Function calls
local message = greet("World");
local sum = add(5, 3);
local time = getCurrentTime();
```

#### Higher-Order Functions and Closures

```atom
// Function that returns a function
func createMultiplier(factor) {
    return func(x) {
        return x * factor;
    };
}

// Closure with captured variables
func createCounter() {
    local count = 0;
    return func() {
        count += 1;
        return count;
    };
}

// Higher-order function
func applyOperation(a, b, operation) {
    return operation(a, b);
}

// Usage
const double = createMultiplier(2);
const counter = createCounter();
const result = applyOperation(5, 3, func(x, y) { return x + y; });
```

#### Recursive Functions

```atom
// Factorial function
func factorial(n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

// Fibonacci with memoization
func fibonacci(n) {
    if (n <= 1) return n;
    return fibonacci(n - 1) + fibonacci(n - 2);
}
```

### Arrays

#### Array Creation and Manipulation

```atom
// Array creation
const empty = [];
const numbers = [1, 2, 3, 4, 5];
const mixed = [1, "hello", true, null, [1, 2, 3]];

// Array indexing
local first = numbers[0];    // 1
local last = numbers[4];     // 5
numbers[1] = 99;             // Modify element

// Array methods
numbers.push(6);             // Add to end
local popped = numbers.pop(); // Remove from end
local length = numbers.length(); // Get length
```

#### Array Functional Methods

```atom
const data = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

// Map (select)
const doubled = data.select(func(index, value) {
    return value * 2;
});

// Filter (where)
const evens = data.where(func(value) {
    return value % 2 == 0;
});

// Iteration (each)
data.each(func(index, value) {
    println("Index:", index, "Value:", value);
});

// All elements satisfy condition
const allPositive = data.all(func(value) {
    return value > 0;
});

// Any elements exist
const hasElements = data.any();
```

### Objects

#### Object Creation and Access

```atom
// Object creation
const person = {
    name: "John",
    age: 30,
    city: "New York"
};

// Property access
local name = person.name;           // Dot notation
local age = person["age"];          // Bracket notation

// Property modification
person.age = 31;
person["city"] = "Boston";
person.newProperty = "value";       // Add new property
```

#### Nested Objects

```atom
const company = {
    name: "TechCorp",
    employees: [
        {
            name: "Alice",
            position: "Developer",
            skills: ["JavaScript", "Python", "Go"],
            contact: {
                email: "alice@techcorp.com",
                phone: "555-0101"
            }
        }
    ],
    departments: {
        engineering: {
            budget: 1000000,
            head: "Charlie"
        }
    }
};

// Access nested properties
local email = company.employees[0].contact.email;
local budget = company.departments.engineering.budget;
```

### Classes and Object-Oriented Programming

#### Class Definition

```atom
class Animal {
    func init(self, name, species) {
        self.name = name;
        self.species = species;
    }

    func speak(self) {
        println(self.name + " makes a sound");
    }

    func getInfo(self) {
        return self.name + " is a " + self.species;
    }
}

// Class instantiation
const dog = new Animal("Buddy", "dog");
dog.speak();
println(dog.getInfo());
```

#### Class with Methods

```atom
class BankAccount {
    func init(self, owner, initialBalance) {
        self.owner = owner;
        self.balance = initialBalance;
        self.transactions = [];
    }

    func deposit(self, amount) {
        self.balance += amount;
        self.transactions.push("Deposit: +" + amount);
    }

    func withdraw(self, amount) {
        if (self.balance >= amount) {
            self.balance -= amount;
            self.transactions.push("Withdrawal: -" + amount);
            return true;
        }
        return false;
    }

    func getBalance(self) {
        return self.balance;
    }
}

// Usage
const account = new BankAccount("Alice", 1000);
account.deposit(500);
account.withdraw(200);
println("Balance:", account.getBalance());
```

### Asynchronous Programming

#### Async/Await

```atom
// Async function declaration
async func fetchData(id) {
    println("Fetching data for ID:", id);
    return "Data-" + id;
}

// Await usage
async func processData() {
    const data1 = await fetchData("001");
    const data2 = await fetchData("002");
    println("All data:", data1, data2);
    return [data1, data2];
}

// Nested async calls
async func complexAsync() {
    const result = await processData();
    const processed = await processResult(result);
    return processed;
}
```

#### Async with Control Flow

```atom
// Async with loops
async func processMultiple() {
    local results = [];
    for (local i = 0; i < 5; i += 1) {
        const data = await fetchData("item-" + i);
        results.push(data);
    }
    return results;
}

// Async with conditionals
async func conditionalAsync(shouldProcess) {
    if (shouldProcess) {
        const result = await fetchData("process");
        return result;
    } else {
        return "skipped";
    }
}
```

### Error Handling

```atom
// Catch expressions for error handling
func riskyOperation() {
    // Potentially failing code with catch expression
    local result = someOperation() catch (error) {
        println("Error occurred:", error);
        return null;
    };
    return result;
}

// Error propagation with catch
func mayFail(shouldFail) {
    if (shouldFail) {
        throw("Operation failed");
    }
    return "Success!";
}

// Catch expression with error handling
func safeOperation() {
    local result = someRiskyFunction() catch (err) {
        std.throw(err);
        return 1;
    };
    return result;
}

// Multiple operations with catch
func processData() {
    local data1 = fetchData("001") catch (err) {
        println("Failed to fetch data1:", err);
        return null;
    };
    
    local data2 = fetchData("002") catch (err) {
        println("Failed to fetch data2:", err);
        return null;
    };
    
    return [data1, data2];
}
```

### Modules and Imports

#### Standard Library Imports

```atom
// Import entire module
import "atom:std";
import "atom:math";

// Import specific functions
import [print, println, throw] from "atom:std";
import [freeze] from "atom:object";

// Usage
println("Hello, World!");
local result = math.pow(2, 3);
const frozen = freeze(someObject);
```

### Scope and Variable Lifecycle

#### Variable Scoping

```atom
var globalVar = "global";

func demonstrateScope() {
    local localVar = "local";
    
    {
        local blockVar = "block";
        println("Inside block:", blockVar);
        println("Local var:", localVar);
        println("Global var:", globalVar);
    }
    
    // blockVar is not accessible here
    println("Local var:", localVar);
    println("Global var:", globalVar);
}
```

#### Closure Scope

```atom
func createClosure() {
    local captured = "captured";
    
    return func() {
        // Can access captured variable
        return captured + " in closure";
    };
}

const closure = createClosure();
println(closure()); // "captured in closure"
```

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

## Getting Started

### Installation

1. Clone the repository:
```bash
git clone https://github.com/HolliShake/atomv3.git
cd atomv3
```

2. Build the project:
```bash
go build -o atom app/main.go
```

3. Run Atom programs:
```bash
./atom examples/hello.atom
```

### Example Programs

#### Hello World
```atom
import [println] from "atom:std";
println("Hello, World!");
```

#### Simple Calculator
```atom
import [println] from "atom:std";

func add(a, b) {
    return a + b;
}

func multiply(a, b) {
    return a * b;
}

println("5 + 3 =", add(5, 3));
println("4 * 7 =", multiply(4, 7));
```

#### Class Example
```atom
import [println] from "atom:std";

class Person {
    func init(self, name, age) {
        self.name = name;
        self.age = age;
    }
    
    func greet(self) {
        return "Hello, I'm " + self.name + " and I'm " + self.age + " years old.";
    }
}

const person = new Person("Alice", 30);
println(person.greet());
```

## Standard Library

Atom comes with a comprehensive standard library:

- **atom:std**: Core functions like `print`, `println`, `throw`
- **atom:math**: Mathematical functions like `pow`, `sqrt`, `abs`, `floor`, `ceil`
- **atom:object**: Object utilities like `freeze`, `keys`
- **atom:os**: Operating system interface
- **atom:path**: Path manipulation utilities

## Language Design Philosophy

Atom is designed with the following principles:

1. **Simplicity**: Clean, readable syntax that's easy to learn
2. **Expressiveness**: Powerful features for complex programming tasks
3. **Performance**: Efficient execution through compilation to bytecode
4. **Extensibility**: Modular design allowing easy addition of new features
5. **Safety**: Built-in error handling and memory management

## Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.