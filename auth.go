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
		fmt.Println("🔑 Please run 'pls login' to connect your account.")
		fmt.Println("If you've already logged in, your keychain might be locked.")
		os.Exit(1)
	}
	return strings.TrimSpace(string(out))
}

func addApiKey() {
	fmt.Println("Please enter your API key from https://console.anthropic.com/settings/keys:")
	var apiKey string
	fmt.Scanln(&apiKey)

	// First try to delete any existing entry (ignore errors if it doesn't exist)
	deleteCmd := exec.Command("security", "delete-generic-password", "-s", "pls_cli_auth", "-a", "user")
	deleteCmd.Run()
	
	// Now add the new API key
	cmd := exec.Command("security", "add-generic-password", "-s", "pls_cli_auth", "-a", "user", "-w", apiKey)
	err := cmd.Run()
	
	if err != nil {
		fmt.Println("")
		fmt.Println("❌ Unable to save API key to keychain.")
		fmt.Println("Run pls login again and enter your computer password if prompted.")
		fmt.Println("If that doesn't work, run \033[34msecurity unlock-keychain\033[0m and try again.")
		fmt.Println("")
		os.Exit(1)
	}
	
	color.Green("✅ Successfully logged in")
}

func removeApiKey() {
	cmd := exec.Command("security", "delete-generic-password", "-s", "pls_cli_auth", "-a", "user")
	err := cmd.Run()
	if err != nil {
		fmt.Println("❌ Error logging out.")
		fmt.Println("The API key might not exist or your keychain might be locked.")
		os.Exit(1)
	}
	color.Green("✅ Successfully logged out.")
}
