package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runSanityCheck() error {
	editedFiles, err := getEditedFiles()
	if err != nil {
		return err
	}
	editedFilesString := ""
	for filePath, fileContent := range editedFiles {
		editedFilesString += filePath + ":\n" + fileContent + "\n\n"
	}
	response, err := getProblems(editedFilesString)
	if err != nil {
		return err
	}
	if response == nil {
		fmt.Println("No problems found in your staged changes")
		return nil
	}
	return runFixer(*response)
}

func runFixer(problems []string) error {
	for index, problem := range problems {
		count := string(index + 1)
		fmt.Println(count + " / " + string(len(problems)))
		fmt.Println(problem)

		fmt.Println("fix - edits your code")
		fmt.Println("ask <question> - ask a follow up about this")
		fmt.Println("return - skip")

		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			return err
		}
		if strings.ToLower(response) == "fix" {
			// Apply the fix
			// Commit the fix with the explaination
		}
		if strings.HasPrefix(strings.ToLower(response), "ask ") {
			// Ask a follow-up question
		}
	}
	fmt.Println("Do you want to merge these changes into a single commit? (y/N)")
	return nil
}

func getProblems(content string) (*[]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "repo root"
	}
	systemPrompt := "You will be provided files that have changed since the user's last commit. carefully analyze each one and look for possible bugs. take note of the larger context and request diffs or other files with your tools. " +
		"\nIf there are no problems, return an empty array. If there are, respond with an array in the following format [{'filePath': 'path/file.ext', 'explaination': '<1 sentance problem recap>', 'suggestion': '<concise explaination on how to fix without code>'}]. Do not output any other text. Your current working directory is: " + cwd
	userPrompt := content
	response := sendWithFunctions(&systemPrompt, &userPrompt, nil)

	var problems []string
	err = json.Unmarshal([]byte(response), &problems)
	if err != nil {
		return nil, err
	}
	return &problems, nil
}

func getDiff() (string, error) {
	cmd := exec.Command("git", "diff")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func getEditedFiles() (map[string]string, error) {
	editedFiles := make(map[string]string)

	// Execute git diff command to get the list of edited files
	cmd := exec.Command("git", "diff", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Split the output by lines to get individual file paths
	filePaths := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Retrieve the content of each edited file
	for _, filePath := range filePaths {
		if filePath != "" {
			cmd := exec.Command("git", "show", "HEAD:"+filePath)
			content, err := cmd.Output()
			if err != nil {
				return nil, err
			}
			editedFiles[filePath] = string(content)
		}
	}

	return editedFiles, nil
}
