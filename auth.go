package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func apiKey() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configPath := filepath.Join(homeDir, ".pls_key")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

func addApiKey() {
	fmt.Println("Please enter your Groq API key from https://console.groq.com/keys:")
	var apiKey string
	fmt.Scanln(&apiKey)

	setApiKey(apiKey)
	color.Green("✅ Successfully logged in")
}
