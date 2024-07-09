package main

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var isLoading bool = false

var rootCmd = &cobra.Command{
	Use:   "ask <command>",
	Short: "GPT in your terminal",
	Long:  `Cut down on the copy pasting and use AI in your terminal.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printHelp()
			os.Exit(0)
		}

		action := args[0]

		if action == "login" {
			addApiKey()
		}

		if action == "fix" {
			// Get the last terminal output and suggest a fix
			// go showProgressWheel()
			// time.Sleep(10 * time.Second)
			suggestFix()
			color.Yellow("Not implemented yet")
		} else if action == "cmd" {
			// Output ONLY a list of shell commands that complete the task
			// Double back and ask it to confirm that it will work. Then paralell summarize each command

			// Functions for this method:
			// Get previous command (up to n commands)
			// Get previous output (up to n lines)

			// Show commands + summaries for each. Execute one by one. Allow for undo on failure
		} else if action == "chat" {
			// Just send the message to the chatbot
		} else if action == "test" {
			// Generate tests for a file
			// Start by passing in just that file
			// Then add context from referenced functions, etc.
			// Either use the LSP for this or get AI to search the vectordb for it...
			color.Yellow("Not implemented yet")
		} else if action == "repo" {
			color.Yellow("Not implemented yet")
		} else if action == "trace" {
			color.Yellow("Not implemented yet")
		} else if action == "refresh" {
			// Regenerate vectordb
			color.Yellow("Not implemented yet")
		} else if action == "z" {
			// Get last command, ask AI for the reverse of it
			color.Yellow("Not implemented yet")
		} else if action == "y" {
			// Go the other direction
			color.Yellow("Not implemented yet")
		} else if action == "logout" {
			removeApiKey()
		} else {
			printHelp()
		}
	},
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
