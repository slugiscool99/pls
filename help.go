package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
)

func printHelp() {
	fmt.Println("")
	printWelcome()
	fmt.Println("")
	printFormattedColumns()
	fmt.Println("")
}
func printFormattedColumns() {
	options := []struct {
		Option      string
		Description string
	}{
		{"pls \033[1mcmd\033[0m \033[34m'description'\033[0m", "Writes shell commands"},
		{"pls \033[1mwrite\033[0m \033[34m'instructions'\033[0m", "Writes regex, code, etc"},
		{"pls \033[1mcheck\033[0m", "Checks your current git diff for issues"},
		{"pls \033[1mupdate\033[0m or \033[1mlogin\033[0m or \033[1mlogout\033[0m", "Manage pls settings"},
		// {"pls \033[1mcode\033[0m \033[34m'prompt'\033[0m", "Writes code snippets"},
		// {"pls \033[1mexplain\033[0m \033[34m'question'\033[0m", "Answers a question about your repo"},
		// {"\033[1mfind\033[0m \033[34m'error message'\033[0m", "Helps diagnose errors"},
		// {"\033[1mupdate\033[0m \033[32mpath/file.ext \033[34m'instructions'\033[0m", "Updates a file, using related code as context"},
		// {"pls \033[1mcheck\033[0m", "Checks your current commit for issues (run after git commit, before git push)"},
		// {"pls \033[1mrefresh\033[0m", "Fixes "},
	}

	// Determine the maximum length of the option strings for alignment
	maxOptionLen := 0
	for _, opt := range options {
		plainOption := stripANSI(opt.Option)
		if len(plainOption) > maxOptionLen {
			maxOptionLen = len(plainOption)
		}
	}

	spaceFromStart := maxOptionLen + 3

	for _, opt := range options {
		space := strings.Repeat(" ", spaceFromStart-len(stripANSI(opt.Option)))
		fmt.Printf("%s%s%s\n", opt.Option, space, opt.Description)
	}
}

func stripANSI(input string) string {
	re := regexp.MustCompile(`\033\[[0-9;]*m`)
	return re.ReplaceAllString(input, "")
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
