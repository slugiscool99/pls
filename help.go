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
	// fmt.Println("\033[1mpls get \033[34m'question'\033[0m\033[0m\x1b[3m Generate code snippets or terminal commands\x1b[0m")
	// fmt.Println("\033[1mpls fix \033[34m'instructions or errors'\033[0m\033[0m\x1b[3m Fix problems in your repo (will edit files)\x1b[0m")
	// fmt.Println("\033[1mpls docs\033[0m or\033[1m comments \033[34m'path/to/file.ext'\033[0m\033[0m\x1b[3m Add docs (commments above functions only) or comments (explainations throughout file) \x1b[0m")
	fmt.Println("\033[1mpls explain \033[34m'question'\033[0m\033[0m\x1b[3m Answers a question about your repo\x1b[0m")
	fmt.Println("\033[1mpls update \033[32mpath/to/file.ext \033[34m'instructions'\033[0m\033[0m\x1b[3m Updates a file, using related code as context\x1b[0m")
	fmt.Println("\033[1mpls check\033[0m\x1b[3m Checks your current commit for issues (run after git commit, before git push)\x1b[0m")
	fmt.Println("")
	fmt.Println("\033[1mpls login\033[0m or\033[1m logout\033[0m\x1b[3m \x1b[0m")
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
