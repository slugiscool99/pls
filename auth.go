package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

func apiKey() string {
	out, err := exec.Command("security", "find-generic-password", "-s", "ask_cli_auth", "-a", "user", "-w").Output()
	if err != nil {
		fmt.Println("Please run 'ask login' to connect your account.")
		os.Exit(1)
	}
	return string(out)
}

func addApiKey() {
	fmt.Println("Please enter your OpenAI API key:")
	var apiKey string
	fmt.Scanln(&apiKey)
	cmd := exec.Command("security", "add-generic-password", "-s", "ask_cli_auth", "-a", "user", "-w", apiKey)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error saving API key.")
		os.Exit(1)
	}
	color.Green("Successfully logged in")
	color.Yellow("Upgrade your plan to use ask in private repos")
}

func removeApiKey() {
	cmd := exec.Command("security", "delete-generic-password", "-s", "ask_cli_auth", "-a", "user")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error logging out.")
		os.Exit(1)
	}
	color.Green("Successfully logged out.")
}
