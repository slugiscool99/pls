package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func reSource() {
	shell := whichShell()
	if shell == "zsh" {
		if err := exec.Command("zsh", "-c", "source ~/.zshrc").Run(); err != nil {
			fmt.Printf("Failed to source ~/.zshrc: %v\n", err)
		}
	} else if shell == "bash" {
		if err := exec.Command("bash", "-c", "source ~/.bashrc").Run(); err != nil {
			fmt.Printf("Failed to source ~/.bashrc: %v\n", err)
		}
	}
}

func whichShell() string {
	shell := os.Getenv("SHELL")

	if strings.Contains(shell, "zsh") {
		return "zsh"
	} else if strings.Contains(shell, "bash") {
		return "bash"
	} else {
		return ""
	}
}
