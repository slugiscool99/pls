package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

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
		query := strings.Join(args[1:], " ")

		if action == "cmd" {
			runCommand(query)
		} else if action == "write" {
			generateAnswer(query)
		} else if action == "explain" {
			explainLastOutput(query)
		} else if action == "check" {
			checkDiff()
		} else if action == "login" {
			addApiKey()
		} else if action == "logout" {
			removeApiKey()
		} else if action == "help" {
			printHelp()
		} else if action == "clear" {
			deleteSavedFiles()
		} else if action == "update" {
			updatePls()
		} else {
			fmt.Println("Unknown command:", action)
		}

		incrementUsage(action)
	},
}

func runCommand(query string) {
	commands := createShellCommand(query, true)
	err := clipboard.WriteAll(commands)
	if err != nil {
		log.Fatalf("Failed to copy to clipboard: %v", err)
	}
	saveLastCommand(commands)
	fmt.Println("\033[3mCopied to clipboard. Run \033[1mpls explain\033[0m\033[3m to describe each step or \033[1mpls explain 'question'\033[0m\033[3m to ask a follow up\033[0m")
	fmt.Println("")
}


func generateAnswer(query string) {
	code := answerQuestion(query)
	err := clipboard.WriteAll(code)
	if err != nil {
		log.Fatalf("Failed to copy to clipboard: %v", err)
	}
	saveLastCommand(code)
	fmt.Println("\033[3mCopied to clipboard. Run \033[1mpls explain\033[0m\033[3m to elaborate or \033[1mpls explain 'question'\033[0m\033[3m to ask a follow up\033[0m")
	fmt.Println("")
}

func explainLastOutput(query string) {
	commands := getLastOutput()
	if commands == "" {
		fmt.Println("No command to explain")
	}
	explainEachLine(commands, query)
	fmt.Println("")
}

func checkDiff() {
	cmd := exec.Command("git", "diff")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to get git diff: %v", err)
	}
	diff := string(output)
	if diff == "" {
		fmt.Println("No changes to check")
		return
	}
	analyzeDiff(diff)
}

func incrementUsage(action string) {
	data := map[string]string{
		"action": action,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	resp, err := http.Post("https://pls.mom/usage", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var respData map[string]interface{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return
	}

	needs_update := respData["needs_update"]
	if needs_update == true {
		fmt.Println("Update required. Please enter your password if needed.")
		updatePls()
	}
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

func saveLastCommand(query string) {
	filePath := filepath.Join(os.Getenv("HOME"), ".pls", "last_command")
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	err := os.WriteFile(filePath, []byte(query), 0644)
	if err != nil {
		fmt.Println("Error saving history. Run \033[1mpls clear\033[0m to reset.", err)
	}
}

func updatePls() {
	cmd := exec.Command("sh", "-c", "curl -s https://pls.mom/install.sh | sudo bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error updating pls:", err)
	}
}

func getLastOutput() string {
	filePath := filepath.Join(os.Getenv("HOME"), ".pls", "last_command")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return string(content)
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
