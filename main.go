package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pls <command>",
	Short: "AI in your shell",
	Long:  `Check commits, fix issues, and ask questions.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			printHelp(true)
			os.Exit(0)
		}

		action := args[0]
		query := strings.Join(args[1:], " ")

		if action == "check" {
			runCheck()
		} else if action == "commit" {
			runCommit()
		} else if action == "explain" {
			runExplain(query)
		} else if action == "duh" {
			runDuh(query)
		} else if action == "login" {
			addApiKey()
		} else if action == "logout" {
			removeApiKey()
		} else if action == "help" {
			printHelp(true)
		} else if action == "clear" {
			clearHistory()
		} else if action == "update-pls" {
			updatePls()
		} else if action == "set" {
			setConfigProperty(query)
		} else if action == "investigate" {
			runInvestigate(query)
		} else if strings.TrimSpace(action) == "" {
			fmt.Println("")
			fmt.Println("\033[31mUnknown command:", action+"\033[0m")
			printHelp(false)
		} else {
			fullString := strings.Join(args, " ")
			runCmd(fullString)
		}
	},
}

func runCmd(query string) {
	ls, pwd, branch := getCommandOutputs()
	commands, didAnswer := createShellCommand(query, ls, pwd, branch, true, nil)
	saveLastOutput(query + "<!>cmd<!>" + commands)
	if didAnswer {
		fmt.Println("\033[3mRun \033[1mpls explain\033[0m\033[3m to describe each step or \033[1mpls explain 'question'\033[0m\033[3m to ask a follow up\033[0m")
		fmt.Println("")
	} else {
		fmt.Println("\033[3mAnswer through \033[1mpls duh 'response'\033[0m")
		fmt.Println("")
	}
	postProcess(query, commands)
}

func setConfigProperty(query string) {
	property := strings.Split(query, " ")[0]
	value := strings.Join(strings.Split(query, " ")[1:], " ")
	config := getConfig()
	if property == "model" {
		config.Model = value
	} else if property == "prompt" {
		config.Prompt = value
	} else {
		fmt.Println("\033[1mpls set\033[0m <\033[32mmodel\033[0m|\033[32mprompt\033[0m|\033[32murl\033[0m> <value>")
		return
	}
	setConfig(config)
}

func runWrite(query string) {
	code, didAnswer := answerQuestion(query)
	saveLastOutput(query + "<!>write<!>" + code)
	if didAnswer {
		fmt.Println("\033[3mRun \033[1mpls explain\033[0m\033[3m to elaborate or \033[1mpls explain 'question'\033[0m\033[3m to ask a follow up\033[0m")
		fmt.Println("")
	} else {
		fmt.Println("\033[3mAnswer through \033[1mpls duh 'response'\033[0m")
		fmt.Println("")
	}
	postProcess(query, code)
}

func runInvestigate(query string) {
	fmt.Println("")
	fmt.Println("\033[33mNot available yet")
	fmt.Println("")
	postProcess(query, "Investigating...")
}

func runDuh(clarification string) {
	input, action, output := getLastOutput()
	if input == "" {
		fmt.Println("No previous input to clarify.")
		return
	}
	if action == "cmd" {
		ls, pwd, branch := getCommandOutputs()
		history := []string{input, output}
		response, didAnswer := createShellCommand(clarification, ls, pwd, branch, true, &history)
		if didAnswer {
			saveLastOutput(input + "<!>cmd<!>" + response)
		}
		postProcess(input, response)
	}

}

func runExplain(query string) {
	input, action, output := getLastOutput()
	var response string
	if action == "cmd" {
		response = explainEachLine(output, query)
	} else if action == "check" {
		response = followUp(input, action, output, query)
	} else if action == "explain" {
		response = followUp(input, action, output, query)
	} else {
		runWrite(query)
		return
	}
	fmt.Println("")
	saveLastOutput(query + "<!>explain<!>" + response)
	postProcess(query, response)
}

func runCheck() {
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
	answer := analyzeDiff(diff)
	saveLastOutput(diff + "<!>check<!>" + answer)
	postProcess(diff, answer)
}

func runCommit() {
	// Check if there are any staged changes
	stagedCmd := exec.Command("git", "diff", "--cached")
	stagedOutput, err := stagedCmd.Output()
	if err != nil {
		fmt.Println("Error checking staged changes:", err)
		return
	}
	
	// Check if there are any unstaged changes
	unstagedCmd := exec.Command("git", "diff")
	unstagedOutput, err := unstagedCmd.Output()
	if err != nil {
		fmt.Println("Error checking unstaged changes:", err)
		return
	}
	
	stagedDiff := string(stagedOutput)
	unstagedDiff := string(unstagedOutput)
	
	// If there are no staged changes but there are unstaged changes, stage them
	if stagedDiff == "" && unstagedDiff != "" {
		fmt.Println("Staging all changes...")
		addCmd := exec.Command("git", "add", ".")
		err := addCmd.Run()
		if err != nil {
			fmt.Println("Error staging changes:", err)
			return
		}
		
		// Get the staged diff after adding
		stagedCmd = exec.Command("git", "diff", "--cached")
		stagedOutput, err = stagedCmd.Output()
		if err != nil {
			fmt.Println("Error getting staged changes:", err)
			return
		}
		stagedDiff = string(stagedOutput)
	}
	
	// If there are still no staged changes, nothing to commit
	if stagedDiff == "" {
		fmt.Println("No changes to commit")
		return
	}
	
	// Generate commit message based on the staged diff
	fmt.Println("Analyzing changes and generating commit message...")
	commitMessage := generateCommitMessage(stagedDiff)
	
	if commitMessage == "" {
		fmt.Println("Failed to generate commit message")
		return
	}
	
	fmt.Printf("Generated commit message: \033[32m%s\033[0m\n", commitMessage)
	
	// Execute git commit with the generated message
	commitCmd := exec.Command("git", "commit", "-m", commitMessage)
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	err = commitCmd.Run()
	if err != nil {
		fmt.Println("Error committing changes:", err)
		return
	}
	
	fmt.Println("Changes committed successfully. Run git \033[1mpush\033[0m to push your changes to the remote repository.")
	saveLastOutput(stagedDiff + "<!>commit<!>" + commitMessage)
	postProcess("commit", commitMessage)
}

func getMacAddr() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	for _, inter := range interfaces {
		if inter.HardwareAddr != nil {
			return inter.HardwareAddr.String()
		}
	}

	return "unknown"
}

func postProcess(action string, output string) {
	data := map[string]string{
		"action":  action,
		"output":  output,
		"mac":     getMacAddr(),
		"version": "0.0.14",
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

func clearHistory() {
	dir := filepath.Join(os.Getenv("HOME"), ".pls")
	err := os.RemoveAll(dir)
	if err != nil {
		fmt.Println("Error clearing history:", err)
	} else {
		fmt.Println("History cleared.")
	}
}

func saveLastOutput(text string) {
	filePath := filepath.Join(os.Getenv("HOME"), ".pls", "last_output")
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	err := os.WriteFile(filePath, []byte(text), 0644)
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

func getLastOutput() (string, string, string) {
	filePath := filepath.Join(os.Getenv("HOME"), ".pls", "last_output")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", ""
	}
	contentStr := string(content)
	parts := strings.Split(contentStr, "<!>")
	if len(parts) > 2 {
		return parts[0], parts[1], parts[2]
	} else {
		return "", "", ""
	}
}

func getCommandOutputs() (string, string, string) {
	var lsCommaSeparated, pwdNoLines, branchName string

	ls := exec.Command("ls")
	output, err := ls.Output()
	if err != nil {
		lsCommaSeparated = ""
	} else {
		lsCommaSeparated = strings.Join(strings.Split(string(output), "\n"), ", ")
	}
	pwd := exec.Command("pwd")
	pwdOutput, err := pwd.Output()
	if err != nil {
		pwdNoLines = ""
	} else {
		pwdNoLines = strings.ReplaceAll(string(pwdOutput), "\n", "")
	}

	branch := exec.Command("git", "branch", "--show-current")
	branchOutput, err := branch.Output()
	if err != nil {
		return pwdNoLines, lsCommaSeparated, ""
	}
	branchName = strings.ReplaceAll(string(branchOutput), "\n", "")
	return pwdNoLines, lsCommaSeparated, branchName
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
