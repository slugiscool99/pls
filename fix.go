package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func readLogFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}

	logFilePath := filepath.Join(homeDir, "terminal_output.log")
	data, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		return "", fmt.Errorf("error reading log file: %v", err)
	}

	return string(data), nil
}
