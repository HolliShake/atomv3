package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
)

func Error(file string, data []rune, message string, position AtomPosition) {
	pc, _, _, ok := runtime.Caller(1) // 1 = caller frame
	caller := ""
	if ok {
		caller = runtime.FuncForPC(pc).Name()
	}

	// Split the data into lines
	content := string(data)
	lines := strings.Split(content, "\n")

	// Ensure we have at least one line
	if len(lines) == 0 {
		lines = []string{""}
	}

	padding := 3
	start := int(math.Max(0, float64(position.LineStart-padding)))
	end := int(math.Min(float64(len(lines)-1), float64(position.LineEnded+padding)))

	// Print error header
	err_message := fmt.Sprintf("DEBUG(%s)::Error in [%s:%d:%d] %s\n", caller, file, position.LineStart, position.ColmStart, message)

	// Display lines with padding
	for i := start; i <= end; i++ {
		lineNum := i + 1 // 1-based line numbering
		line := lines[i]

		// Format line number with padding
		err_message += fmt.Sprintf("%4d | %s\n", lineNum, line)

		// Add error highlighting if this is the error line
		if i >= position.LineStart && i <= position.LineEnded {

			// Add carets to show the exact error position
			if i == position.LineStart {
				// Show column range for the error
				colStart := position.ColmStart - 1 // Convert to 0-based indexing
				colEnd := position.ColmEnded - 1   // Convert to 0-based indexing

				// Ensure column positions are within bounds
				if colStart < 0 {
					colStart = 0
				}
				if colEnd >= len(line) {
					colEnd = len(line) - 1
				}
				if colEnd < colStart {
					colEnd = colStart
				}

				// Create the error indicator line
				errorLine := strings.Repeat(" ", colStart)
				errorLine += strings.Repeat("^", colEnd-colStart+1)
				err_message += fmt.Sprintf("%4s | %s\n", "", errorLine)
			}
		}
	}

	fmt.Print(err_message)
	os.Exit(1)
}
