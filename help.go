package main

import (
	"fmt"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
)

func printHelp() {
	fmt.Println("")
	printWelcome()
	fmt.Println("")
	fmt.Println("\033[1mdo sh \033[34m'task'\033[0m\033[0m\x1b[3m create shell commands\x1b[0m")
	fmt.Println("\033[1mdo e \033[34m'error'\033[0m\033[0m\x1b[3m investigate an error message\x1b[0m")
	fmt.Println("\033[1mdo q \033[34m'question'\033[0m\033[0m\x1b[3m get quick answers\x1b[0m")
	fmt.Println("\033[1mdo check\033[0m\x1b[3m checks your current commit\x1b[0m")
	fmt.Println("\033[1mdo login\033[0m or\033[1m logout\033[0m\x1b[3m \x1b[0m")
	fmt.Println("")
	// fmt.Println("Available soon:")
	// fmt.Println("")
	// fmt.Println("\033[1mdo repo \033[34myour question\033[0m\x1b[3m do a question about your entire codebase\x1b[0m")
	// fmt.Println("\033[1mdo test \033[34m<input_file> <output_test_file>\033[0m\x1b[3m generate tests for a file\x1b[0m")
	// fmt.Println("\033[1mdo trace \033[34mfirestore|sql|<url>\033[0m\x1b[3m find code that interacts with a resource\x1b[0m")
	// fmt.Println("\033[1mdo z\033[0m\x1b[3m undo command\x1b[0m")
	// fmt.Println("\033[1mdo y\033[0m\x1b[3m redo command\x1b[0m")
}

func printWelcome() {
	// Create ASCII art
	myFigure := figure.NewFigure("do", "puffy", true)

	// Convert ASCII art to string and split by lines
	asciiArt := myFigure.String()
	lines := splitLines(asciiArt)

	// Print each line in a different color
	colors := []func(format string, a ...interface{}){
		color.Cyan,
		color.Magenta,
		color.Yellow,
		color.Green,
		color.Blue,
		color.Red,
	}

	for i, line := range lines {
		colors[i%len(colors)](line)
	}
}

// Helper function to split a string into lines
func splitLines(s string) []string {
	var lines []string
	var currentLine string
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(r)
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}
