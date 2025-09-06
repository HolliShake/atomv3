package main

import (
	"fmt"
	"os"
	"strings"

	runtime "dev.runtime"
)

func readFile(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(content)
}

func testErrorDisplay() {
	// Test the error display functionality
	code := readFile("./test.atom")

	// Test 1: Single line error
	fmt.Println("Test 1: Single line error")
	position1 := Position{
		LineStart: 1, // 0-based line number (line 2)
		LineEnded: 1, // Same line
		ColmStart: 5, // 0-based column
		ColmEnded: 10,
	}
	Error("test.atom", []rune(code), "Unexpected character sequence", position1)

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Test 2: Multi-line error
	fmt.Println("Test 2: Multi-line error")
	position2 := Position{
		LineStart: 5, // 0-based line number (line 6)
		LineEnded: 7, // 0-based line number (line 8)
		ColmStart: 0, // 0-based column
		ColmEnded: 4,
	}
	Error("test.atom", []rune(code), "Function declaration syntax error", position2)

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Test 3: Error at the beginning of file
	fmt.Println("Test 3: Error at beginning of file")
	position3 := Position{
		LineStart: 0, // 0-based line number (line 1)
		LineEnded: 0, // Same line
		ColmStart: 0, // 0-based column
		ColmEnded: 2,
	}
	Error("test.atom", []rune(code), "Invalid comment syntax", position3)
}

func main() {
	// Test error display functionality
	// fmt.Println("Testing Error Display:")
	// fmt.Println("=====================")
	// testErrorDisplay()

	// Uncomment the following section to also run the tokenizer

	// fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// Test JavaScript-style tokenizer with Unicode support
	code := readFile("./test.atom")

	state := runtime.NewAtomState()

	tokenizer := NewTokenizer("test.atom", code)
	parser := NewParser(tokenizer)
	compile := NewCompile(parser, state)
	i := runtime.NewInterpreter(state)
	i.Interpret(compile.Compile())
}
