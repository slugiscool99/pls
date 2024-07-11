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
	fmt.Println("\033[1mpls \033[34m'anything'\033[0m\033[0m\x1b[3m Try pasting error messages, asking questions, or generating commands\x1b[0m")
	fmt.Println("\033[1mpls fix \033[34m'instructions or errors'\033[0m\033[0m\x1b[3m Fix problems\x1b[0m")
	fmt.Println("\033[1mpls check\033[0m\x1b[3m Checks your current commit (run this after you commit, but before you push)\x1b[0m")
	fmt.Println("")
	fmt.Println("\033[1mpls login\033[0m or\033[1m logout\033[0m\x1b[3m \x1b[0m")
	fmt.Println("")
}

func printWelcome() {
	// Create ASCII art
	myFigure := figure.NewFigure("pls", "puffy", true)

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
