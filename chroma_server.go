package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"
)

func startServer() {
	cmd := exec.Command("pipx", "--version")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Installing pipx...")
		if err := exec.Command("brew", "install", "pipx").Run(); err != nil {
			fmt.Printf("Failed to install pipx: %v\n", err)
			return
		}
		if err := exec.Command("pipx", "ensurepath").Run(); err != nil {
			fmt.Printf("Failed to run pipx ensurepath: %v\n", err)
			return
		}

		reSource()
	}

	cmd = exec.Command("chroma", "--help")
	if _, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("Installing chroma...")
		if err := exec.Command("pipx", "install", "chromadb").Run(); err != nil {
			fmt.Printf("Failed to install chroma: %v\n", err)
			return
		}
	}

	// Check if the directory ~/.pls/db exists
	cmd = exec.Command("ls", "~/.pls/db")
	if _, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("Creating directory ~/.pls/db...")
		if err := exec.Command("mkdir", "-p", "~/.pls/db").Run(); err != nil {
			fmt.Printf("Failed to create directory ~/.pls/db: %v\n", err)
			return
		}
	}

	fmt.Println("Killing all processes on port 8000...")
	killAllProcessesOnPort("8000")

	// Run chroma
	go func() {
		if err := exec.Command("chroma", "run", "--path", "~/.pls/db").Run(); err != nil {
			fmt.Printf("Failed to run chroma: %v\n", err)
		}
	}()

	color.Green("Set up successful")
	go handleCleanup()
}

func stopChroma() error {
	cmd := exec.Command("pkill", "chroma")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error terminating chroma processes: %v", err)
	}
	return nil
}

func handleCleanup() {
	// Create a channel to listen for interrupt or terminate signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sigChan

	// Cleanup actions
	if err := stopChroma(); err != nil {
		fmt.Printf("error during cleanup: %v\n", err)
	}

	fmt.Println("Cleanup complete.")
}

func killAllProcessesOnPort(port string) error {
	out, err := exec.Command("lsof", "-i", ":8000").Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return nil
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return nil
	}
	pid := fields[1]

	// Terminate the process using the PID
	if err := exec.Command("kill", "-9", pid).Run(); err != nil {
		fmt.Printf("Error terminating process with PID %s: %v\n", pid, err)
		return err
	}

	fmt.Printf("Process %s on port %s was terminated\n", pid, port)
	return nil
}
