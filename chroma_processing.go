package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	ignore "github.com/sabhiram/go-gitignore"
)

func processRepo() {
	repoRoot, err := getRepoRoot()
	if err != nil {
		fmt.Printf("error getting repo root: %v", err)
		return
	}

	fmt.Println("")
	fmt.Println("Processing all files except those in .gitignore.")
	color.Yellow("It is recommended to ignore dependency folders like /node_modules, /vendor, etc.")
	fmt.Println("Add these folders to .gitignore if needed and press Enter to continue (type cancel to exit).")
	var continueString string
	fmt.Scanln(&continueString)
	if continueString == "cancel" {
		return
	}

	gitignorePath := filepath.Join(repoRoot, ".gitignore")
	ignoreMatcher, err := ignore.CompileIgnoreFile(gitignorePath)
	if err != nil {
		fmt.Printf("error parsing .gitignore: %v\n", err)
		return
	}

	filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}

		if ignoreMatcher.MatchesPath(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			fmt.Printf("Processing file: %s\n", relPath)
			contents, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("error reading file %s: %v\n", path, err)
				return nil
			}
			processFile(string(contents))
		}

		return nil
	})

}

func processFile(contents string) []string {
	// This is a simplistic approach and might need adjustments based on the actual language syntax
	funcRegex := regexp.MustCompile(`(?m)^(func|def|void|function|sub|public|private|protected|static|procedure)\s+\w+.*\{`)

	matches := funcRegex.FindAllStringIndex(contents, -1)
	if len(matches) == 0 {
		// If no matches found, return the entire content as the only element in the array
		return []string{contents}
	}

	result := []string{}
	start := 0
	nestingLevel := 0

	for i, match := range matches {
		if start < match[0] && nestingLevel == 0 {
			result = append(result, contents[start:match[0]])
		}
		end := match[1]
		if i < len(matches)-1 {
			nextMatch := matches[i+1]
			// Find the end of the current function by looking for the closing brace
			end = findClosingBrace(contents, match[1], nextMatch[0], &nestingLevel)
		} else {
			// Last match, find the closing brace till the end of the file
			end = findClosingBrace(contents, match[1], len(contents), &nestingLevel)
		}
		if nestingLevel == 0 {
			result = append(result, contents[match[0]:end])
			start = end
		}
	}

	if start < len(contents) {
		result = append(result, contents[start:])
	}

	return result
}

func findClosingBrace(contents string, start, end int, nestingLevel *int) int {
	for i := start; i < end; i++ {
		if contents[i] == '{' {
			*nestingLevel++
		} else if contents[i] == '}' {
			*nestingLevel--
			if *nestingLevel == 0 {
				return i + 1
			}
		}
	}
	return end
}

func getRepoRoot() (string, error) {
	// Execute the git command to get the top-level directory
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error executing git command: %w", err)
	}

	// Trim the output to remove any trailing newline characters
	repoRoot := strings.TrimSpace(string(out))

	// Get the absolute path
	absRepoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %w", err)
	}

	return absRepoRoot, nil
}
