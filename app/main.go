package main

import (
	"fmt"
	"os"
	"strings"
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
		LineStart:   1, // 0-based line number (line 2)
		LineEnded:   1, // Same line
		ColumnStart: 5, // 0-based column
		ColumnEnded: 10,
	}
	Error("test.atom", []rune(code), "Unexpected character sequence", position1)

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Test 2: Multi-line error
	fmt.Println("Test 2: Multi-line error")
	position2 := Position{
		LineStart:   5, // 0-based line number (line 6)
		LineEnded:   7, // 0-based line number (line 8)
		ColumnStart: 0, // 0-based column
		ColumnEnded: 4,
	}
	Error("test.atom", []rune(code), "Function declaration syntax error", position2)

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Test 3: Error at the beginning of file
	fmt.Println("Test 3: Error at beginning of file")
	position3 := Position{
		LineStart:   0, // 0-based line number (line 1)
		LineEnded:   0, // Same line
		ColumnStart: 0, // 0-based column
		ColumnEnded: 2,
	}
	Error("test.atom", []rune(code), "Invalid comment syntax", position3)
}

func main() {
	// Test error display functionality
	fmt.Println("Testing Error Display:")
	fmt.Println("=====================")
	testErrorDisplay()

	// Uncomment the following section to also run the tokenizer
	/*
		fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

		// Test JavaScript-style tokenizer with Unicode support
		code := readFile("./test.atom")

		tokenizer := NewTokenizer("test.js", code)

		fmt.Println("Tokenizing JavaScript code with Unicode support:")
		fmt.Println("==================================================")

		for {
			token := tokenizer.NextToken()
			fmt.Printf("%s\n", token.String())

			if token.ttype == TokenTypeEof {
				break
			}
		}
	*/
}
