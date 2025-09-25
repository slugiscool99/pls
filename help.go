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
	fmt.Println("Usage:")
	fmt.Println("  pls <command>           Ask Groq AI anything")
	fmt.Println("  pls login [api-key]     Set your Groq API key")
	fmt.Println("  pls model [model-name]  Set or view the AI model (default: openai/gpt-oss-20b)")
	fmt.Println("  pls help                Show this help message")
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
