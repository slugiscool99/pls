package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

func apiKey() string {
	out, err := exec.Command("security", "find-generic-password", "-s", "pls_cli_auth", "-a", "user", "-w").Output()
	if err != nil {
		fmt.Println("Please run 'pls login' to connect your account.")
		os.Exit(1)
	}
	return strings.TrimSpace(string(out))
}

func addApiKey() {
	fmt.Println("Please enter your API key from https://console.groq.com/keys:")
	var apiKey string
	fmt.Scanln(&apiKey)
	cmd := exec.Command("security", "add-generic-password", "-s", "pls_cli_auth", "-a", "user", "-w", apiKey)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error saving API key.")
		os.Exit(1)
	}
	color.Green("Successfully logged in")
}

func removeApiKey() {
	cmd := exec.Command("security", "delete-generic-password", "-s", "pls_cli_auth", "-a", "user")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error logging out.")
		os.Exit(1)
	}
	color.Green("Successfully logged out.")
}
