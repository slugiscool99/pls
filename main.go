package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var isLoading bool = false

var rootCmd = &cobra.Command{
	Use:   "pls <command>",
	Short: "AI in your shell",
	Long:  `Check commits, fix issues, and ask questions.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			printHelp()
			os.Exit(0)
		}

		action := args[0]
		query := strings.Join(args[0:], " ")

		if action == "login" {
			addApiKey()
		} else if action == "logout" {
			removeApiKey()
		} else if action == "check" {
			runSanityCheck()
		} else if action == "help" {
			printHelp()
		} else if action == "it" {
			runSavedCommand()
		} else if action == "a" {
			askFollowUp()
		} else if action == "clear" {
			deleteSavedFiles()
		} else {
			parseQuery(query)
		}
	},
}

func parseQuery(query string) {

	queryType := checkQueryType(query)
	// intent := checkIntent(query)

	// Follow up logic
	lastQueryType, lastSeenString := getLastCommand()
	saveLastCommand(queryType)

	if lastQueryType != "" && lastSeenString != "" {
		lastSeen, err := time.Parse(time.RFC3339, lastSeenString)
		if err != nil {
			lastSeen = time.Now().Add(-10 * time.Minute)
		}
		if queryType == lastQueryType { // Same type of query
			if time.Now().Sub(lastSeen) < 12*time.Hour { // It's been longer than 1 minute
				askFollowUp(query)
				return
			}
		}
	}

	if queryType == "ERROR" {
		response := checkError(query)
		saveResponse(response, "last.err")
	} else if queryType == "CODE_QUESTION" {
		response := askQuestion(query, "code")
		saveResponse(response, "last.q")
	} else if queryType == "FILE_QUESTION" {
		response := askQuestion(query, "file")
		saveResponse(response, "last.q")
	} else if queryType == "OTHER_QUESTION" {
		response := askQuestion(query, "other")
		saveResponse(response, "last.q")
	} else if queryType == "PROJECT_QUESTION" {
		color.Yellow("Not yet supported")
	} else if queryType == "SHELL_TASK" {
		if strings.TrimSpace(query) == "" {
			printSavedCommand()
			return
		}
		response := createShellCommand(query)
		saveResponse(response, "last.sh")
	} else {
		fmt.Println("I'm not sure what you're asking. Try pasting an error message, asking a question, or generating a command.")
	}

}

func askFollowUp(query string) {
	timeNow

	//follow up
}

func deleteSavedFiles() {
	dir := filepath.Join(os.Getenv("HOME"), ".pls")
	err := os.RemoveAll(dir)
	if err != nil {
		fmt.Println("Error clearing history:", err)
	} else {
		fmt.Println("History cleared.")
	}
}

func saveLastCommand(queryType string) {
	filePath := filepath.Join(os.Getenv("HOME"), ".pls", "last_type")
	timeNow := time.Now().Format(time.RFC3339)
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	err := os.WriteFile(filePath, []byte(queryType+"\n"+timeNow), 0644)
	if err != nil {
		fmt.Println("Error saving history. Run \033[1mpls clear\033[0m to reset.", err)
	}
}

func getLastCommand() (string, string) {
	filePath := filepath.Join(os.Getenv("HOME"), ".pls", "last_type")

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Couldn't find the command to run", err)
		return "", ""
	}
	pieces := strings.Split(string(content), "\n")
	if len(pieces) != 2 {
		return "", ""
	}
	return pieces[0], pieces[1]
}

func saveResponse(response string, file string) {
	filePath := filepath.Join(os.Getenv("HOME"), ".pls", file)
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	err := os.WriteFile(filePath, []byte(response), 0644)
	if err != nil {
		fmt.Println("Error saving history. Run \033[1mpls clear\033[0m to reset.", err)
	} else {
		if file == "last.sh" {
			fmt.Println("Run \033[1mpls it\033[0m to execute")
		} else if file == "last.q" {
			fmt.Println("Run \033[1mpls q\033[0m ask a follow up")
		}
	}
}

func printSavedCommand() {
	filePath := filepath.Join(os.Getenv("HOME"), ".pls", "last.sh")
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Couldn't find the command to run", err)
		return
	}
	fmt.Println(string(content))
}

func runSavedCommand() {
	filePath := filepath.Join(os.Getenv("HOME"), ".pls", "last.sh")
	filePathTiming := filepath.Join(os.Getenv("HOME"), ".pls", "last_seen.sh")

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Couldn't find the command to run", err)
		return
	}

	confirmRunAfterOneMinute(filePathTiming, "Are you sure you want to run this? (y/N): "+string(content))

	cmd := exec.Command("sh", "-c", string(content))
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(output))
}

func confirmRunAfterOneMinute(filePath string, prompt string) {
	mustConfirm := true
	timeNow := time.Now().Format(time.RFC3339)
	oldTime, err := os.ReadFile(filePath)
	if err == nil {
		oldTimeParse, err := time.Parse(time.RFC3339, string(oldTime))
		if err == nil {
			if oldTimeParse.Add(1 * time.Minute).After(time.Now()) {
				mustConfirm = false
			}
		}
	}

	if mustConfirm {
		fmt.Println(prompt)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			return
		}
	}
	_ = os.WriteFile(filePath, []byte(timeNow), 0644)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func showProgressWheel() {
	wheel := []rune{'|', '/', '-', '\\'}
	for {
		for _, r := range wheel {
			fmt.Printf("\r%c", r)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
