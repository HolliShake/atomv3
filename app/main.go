package main

import (
	"fmt"
	"os"
	"path/filepath"
	gruntime "runtime"
	"strings"

	runtime "dev.runtime"
)

func readFile(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return string(content)
}

func runTests(testFile string) {
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	execDir := filepath.Dir(execPath)
	testsDir := filepath.Join(execDir, "test")

	// If a specific test file is provided
	if testFile != "" {
		fileDir := filepath.Join(testsDir, testFile)

		// Add .atom extension if not present
		if !strings.HasSuffix(testFile, ".atom") {
			fileDir += ".atom"
		}

		// Check if file exists
		if _, err := os.Stat(fileDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Test file %s not found\n", testFile)
			os.Exit(1)
		}

		runFile(fileDir)
		return
	}

	// Run all tests in the directory
	files, err := os.ReadDir(testsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	success := 0
	total := 0

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".atom") {
			continue
		}

		total++
		testPath := filepath.Join(testsDir, file.Name())

		// We could add error handling here to continue testing even if one test fails
		runFile(testPath)
		success++
	}

	fmt.Printf("Success: %d Total: %d\n", success, total)
}

func printStartupBanner() {
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                                              ║")
	fmt.Println("║    █████╗ ████████╗ ██████╗ ███╗   ███╗    ██████╗ ██████╗  ██████╗ ███████╗ ║")
	fmt.Println("║   ██╔══██╗╚══██╔══╝██╔═══██╗████╗ ████║   ██╔════╝██╔═══██╗██╔═══██╗██╔════╝ ║")
	fmt.Println("║   ███████║   ██║   ██║   ██║██╔████╔██║   ██║     ██║   ██║██║   ██║███████╗ ║")
	fmt.Println("║   ██╔══██║   ██║   ██║   ██║██║╚██╔╝██║   ██║     ██║   ██║██║   ██║██╔════╝ ║")
	fmt.Println("║   ██║  ██║   ██║   ╚██████╔╝██║ ╚═╝ ██║   ╚██████╗╚██████╔╝╚██████╔╝███████╗ ║")
	fmt.Println("║   ╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝     ╚═╝    ╚═════╝ ╚═════╝  ╚═════╝ ╚══════╝ ║")
	fmt.Println("║                                                                              ║")
	fmt.Println("║                    A Custom Programming Language                             ║")
	fmt.Println("║                    Implemented in Go                                         ║")
	fmt.Println("║                                                                              ║")
	fmt.Println("║  Features: Dynamic Typing • OOP • Functions • Arrays • Objects • Classes     ║")
	fmt.Println("║  Author:   Philipp Andrew Redondo                                            ║")
	fmt.Println("║  License:  MIT License                                                       ║")
	fmt.Println("║  GitHub:   https://github.com/HolliShake/atomv3                              ║")
	fmt.Printf("║  Version:  %s                                                             ║\n", VERSION)
	fmt.Println("║                                                                              ║")
	fmt.Println("║  usage: atom [<file.atom> | --test]                                          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
}

func runFile(file string) {
	code := readFile(file)
	s := runtime.NewAtomState()
	t := NewAtomTokenizer(file, code)
	p := NewAtomParser(t)
	c := NewAtomCompile(p, s)
	f := c.Compile()
	i := runtime.NewInterpreter(s)
	i.Interpret(f)
}

func main() {
	if len(os.Args) < 2 {
		printStartupBanner()
		os.Exit(1)
	}

	if os.Args[1] == "--test" {
		testFile := ""
		if len(os.Args) > 2 {
			testFile = os.Args[2]
		}
		runTests(testFile)
		os.Exit(0)
	}

	gruntime.GC()
	var mStart, mEnd gruntime.MemStats
	absPath, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	runFile(absPath)

	gruntime.ReadMemStats(&mEnd)
	fmt.Printf("💾 Memory usage: %d kilobytes\n", (mEnd.Alloc-mStart.Alloc)/1024)
}
