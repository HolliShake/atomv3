package main

import (
	"fmt"
	"os"
	gruntime "runtime"

	runtime "dev.runtime"
)

func readFile(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(content)
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
	fmt.Println("║                                                                              ║")
	fmt.Println("║  usage: atom <file.atom>                                                     ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
}

func main() {
	if len(os.Args) < 2 {
		printStartupBanner()
		os.Exit(1)
	}

	gruntime.GC()
	var mStart, mEnd gruntime.MemStats
	code := readFile(os.Args[1])
	s := runtime.NewAtomState()
	t := NewAtomTokenizer(os.Args[1], code)
	p := NewAtomParser(t)
	c := NewAtomCompile(p, s)
	f := c.Compile()
	i := runtime.NewInterpreter(s)
	i.Interpret(f)

	gruntime.ReadMemStats(&mEnd)
	fmt.Printf("💾 Memory usage: %d kilobytes\n", (mEnd.Alloc-mStart.Alloc)/1024)
}
