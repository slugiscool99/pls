package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var isLoading bool = false

var rootCmd = &cobra.Command{
	Use:   "do <command>",
	Short: "GPT in your terminal",
	Long:  `Cut down on the copy pasting and use AI in your terminal.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			printHelp()
			os.Exit(0)
		}

		action := args[0]
		query := strings.Join(args[1:], " ")

		if action == "login" {
			addApiKey()
		}
		if action == "check" {
			runSanityCheck()
		} else if action == "e" {
			response := checkError(query)
			saveResponse(response, "last.err")
		} else if action == "q" {
			response := askQuestion(query)
			saveResponse(response, "last.q")
		} else if action == "qq" {
			askFollowUp(query)
		} else if action == "sh" {
			if strings.TrimSpace(query) == "" {
				printSavedCommand()
				return
			}
			response := createShellCommand(query)
			saveResponse(response, "last.sh")
		} else if action == "it" {
			runSavedCommand()
		} else if action == "clear" {
			deleteSavedFiles()
		} else if action == "make" {
			// if len(args) < 2 {
			// 	fmt.Println("Please provide a file path.")
			// 	return
			// }
			// filename := args[1]
			// fmt.Println("Additional instructions (optional):")
			// var instructions string
			// fmt.Scanln(&instructions)
			// if _, err := os.Stat(filename); err == nil {
			// 	fmt.Println("That file already exists. Overwrite? (y/N)")
			// 	var overwrite string
			// 	fmt.Scanln(&overwrite)
			// 	if strings.ToLower(overwrite) != "y" {
			// 		return
			// 	}
			// } else {
			// 	fmt.Println("Error checking file existence:", err)
			// 	return
			// }
		} else if action == "test" {
			testFunctions()
		} else if action == "logout" {
			removeApiKey()
		} else {
			printHelp()
		}
	},
}

func deleteSavedFiles() {
	dir := filepath.Join(os.Getenv("HOME"), ".do")
	err := os.RemoveAll(dir)
	if err != nil {
		fmt.Println("Error clearing history:", err)
	} else {
		fmt.Println("History cleared.")
	}
}

func saveResponse(response string, file string) {
	filePath := filepath.Join(os.Getenv("HOME"), ".do", file)
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	err := os.WriteFile(filePath, []byte(response), 0644)
	if err != nil {
		fmt.Println("Error saving history. Run \033[1mdo clear\033[0m to reset.", err)
	} else {
		if file == "last.sh" {
			fmt.Println("Run \033[1mdo it\033[0m to execute")
		} else if file == "last.q" {
			fmt.Println("Run \033[1mdo qq\033[0m to ask a follow up")
		}
	}
}

func printSavedCommand() {
	filePath := filepath.Join(os.Getenv("HOME"), ".do", "last.sh")
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Couldn't find the command to run", err)
		return
	}
	fmt.Println(string(content))
}

func runSavedCommand() {
	filePath := filepath.Join(os.Getenv("HOME"), ".do", "last.sh")
	filePathTiming := filepath.Join(os.Getenv("HOME"), ".do", "last_seen.sh")

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Couldn't find the command to run", err)
		return
	}

	confirmRun(filePathTiming, "Are you sure you want to run this? (y/N): "+string(content))

	cmd := exec.Command("sh", "-c", string(content))
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(output))
}

func askFollowUp(query string) {
	filePath := filepath.Join(os.Getenv("HOME"), ".do", "last.q")
	filePathTiming := filepath.Join(os.Getenv("HOME"), ".do", "last_seen.q")
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Couldn't find the command to run", err)
		return
	}

	firstFewWords := strings.Join(strings.Split(string(content), " ")[:5], " ")
	confirmRun(filePathTiming, "Did you mean to respond to"+firstFewWords+"...")

	//follow up
}

func confirmRun(filePath string, prompt string) {
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
