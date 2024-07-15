package main

import (
	"bytes"
	"fmt"
	"log"
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
		} else if action == "echo" {
			generateAnswer(query)
		} else if action == "explain" {
			explainLastOutput(query)
		} else if action == "check" {
			runSanityCheck()
		} else if action == "test" {
			setupChroma()
		} else if action == "login" {
			addApiKey()
		} else if action == "logout" {
			removeApiKey()
		} else if action == "help" {
			printHelp()
		} else if action == "clear" {
			deleteSavedFiles()
		} else {
			explainRepo(query)
		}
	},
}

func runCommand(query string) {
	commands := createShellCommand(query, true)
	err := clipboard.WriteAll(commands)
	if err != nil {
		log.Fatalf("Failed to copy to clipboard: %v", err)
	}
	saveLastCommand(commands)
	checkForInstallationNeeds(commands)
	fmt.Println("Copied to clipboard. Run \033[1mpls explain\033[0m to describe each step")
	fmt.Println("")
}

func checkForInstallationNeeds(commands string) {
	needsInstallation := []string{}
	commandList := strings.Split(commands, "\n")
	for _, command := range commandList {
		packageName := strings.Split(command, " ")[0]
		if !isInstalled(packageName) {
			needsInstallation = append(needsInstallation, packageName)
		}
	}

	if len(needsInstallation) > 0 {
		fmt.Print("You may need the following packages: ")
		for _, packageName := range needsInstallation {
			fmt.Print(packageName)
			if packageName != needsInstallation[len(needsInstallation)-1] {
				fmt.Print(", ")
			}
			fmt.Println("")
		}
		fmt.Print("Install them? (y/n): ")
		var input string
		fmt.Scanln(&input)
		if input == "y" {
			installPackages(needsInstallation)
			reSource()
		}
		fmt.Println("")
	}
}

func isInstalled(packageName string) bool {
	cmd := exec.Command("which", packageName)
	err := cmd.Run()
	return err == nil
}

func installPackages(packages []string) {
	for _, packageName := range packages {
		cmd := exec.Command("brew", "install", packageName)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			if bytes.Contains(stderr.Bytes(), []byte("No available formula with the name")) {
				instructions := askForInstallationInstructions(packageName)
				cmd := exec.Command("sh", "-c", instructions)
				err := cmd.Run()
				if err != nil {
					fmt.Println("Error installing package", packageName)
				}
			}
			fmt.Println("Error installing package", packageName)
		}
	}
}

func askForInstallationInstructions(packageName string) string {
	return createShellCommand("brew install "+packageName+": No available formula with the name. What are the shell commands to install "+packageName, false)
}

func explainRepo(query string) {

}

func generateAnswer(query string) {
	lang := getLikelyLanguage()
	code := answerQuestion(query, lang)
	err := clipboard.WriteAll(code)
	if err != nil {
		log.Fatalf("Failed to copy to clipboard: %v", err)
	}
	saveLastCommand(code)
	fmt.Println("Copied to clipboard. Run \033[1mpls explain\033[0m to elaborate")
	fmt.Println("")
}

func explainLastOutput(query string) {
	commands := getLastOutput()
	if commands == "" {
		fmt.Println("No command to explain")
	}
	explainEachLine(commands, query)
}

func getLikelyLanguage() string {
	//not implemented
	return ""
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
