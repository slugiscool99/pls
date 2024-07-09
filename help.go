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
	fmt.Println("\033[1mask fix\033[0m\x1b[3m suggest a fix based on the last terminal output\x1b[0m")
	fmt.Println("\033[1mask make \033[34mfilename\033[0m\x1b[3m generate a file from surrounding context\x1b[0m")
	fmt.Println("\033[1mask cmd \033[34myour task\033[0m\x1b[3m generate one or more shell commands\x1b[0m")
	fmt.Println("\033[1mask chat \033[34myour message\033[0m\x1b[3m send a vanilla GPT query\x1b[0m")
	fmt.Println("")
	fmt.Println("\033[1mask login\033[0m or\033[1m logout\033[0m\x1b[3m manage your account\x1b[0m")
	// fmt.Println("")
	// fmt.Println("Available soon:")
	// fmt.Println("")
	// fmt.Println("\033[1mask repo \033[34myour question\033[0m\x1b[3m ask a question about your entire codebase\x1b[0m")
	// fmt.Println("\033[1mask test \033[34m<input_file> <output_test_file>\033[0m\x1b[3m generate tests for a file\x1b[0m")
	// fmt.Println("\033[1mask trace \033[34mfirestore|sql|<url>\033[0m\x1b[3m find code that interacts with a resource\x1b[0m")
	// fmt.Println("\033[1mask z\033[0m\x1b[3m undo command\x1b[0m")
	// fmt.Println("\033[1mask y\033[0m\x1b[3m redo command\x1b[0m")
}

func printWelcome() {
	// Create ASCII art
	myFigure := figure.NewFigure("ask", "puffy", true)

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
