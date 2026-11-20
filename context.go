package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const maxListedEntries = 25

// Files whose presence says what kind of project the working directory is.
var projectMarkers = []struct {
	file  string
	label string
}{
	{"package.json", "Node.js (package.json)"},
	{"pnpm-lock.yaml", "pnpm"},
	{"yarn.lock", "Yarn"},
	{"bun.lockb", "Bun"},
	{"bun.lock", "Bun"},
	{"package-lock.json", "npm"},
	{"next.config.js", "Next.js"},
	{"next.config.mjs", "Next.js"},
	{"next.config.ts", "Next.js"},
	{"app.json", "Expo/React Native (app.json)"},
	{"go.mod", "Go"},
	{"Cargo.toml", "Rust (Cargo)"},
	{"pyproject.toml", "Python (pyproject.toml)"},
	{"requirements.txt", "Python (requirements.txt)"},
	{"Gemfile", "Ruby (Bundler)"},
	{"Podfile", "CocoaPods"},
	{"Package.swift", "Swift Package Manager"},
	{"build.gradle", "Gradle"},
	{"build.gradle.kts", "Gradle"},
	{"pom.xml", "Maven"},
	{"CMakeLists.txt", "CMake"},
	{"Makefile", "Make"},
	{"Dockerfile", "Docker"},
	{"docker-compose.yml", "Docker Compose"},
	{"compose.yaml", "Docker Compose"},
}

// getWorkingContext returns a short description of the current directory
// (path, shell, git branch, project type, top-level entries) to help the
// model write commands that fit it.
func getWorkingContext() string {
	var lines []string

	if cwd, err := os.Getwd(); err == nil {
		lines = append(lines, "Directory: "+cwd)
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		lines = append(lines, "Shell: "+filepath.Base(shell))
	}
	if out, err := exec.Command("git", "branch", "--show-current").Output(); err == nil {
		if branch := strings.TrimSpace(string(out)); branch != "" {
			lines = append(lines, "Git branch: "+branch)
		}
	}

	entries, err := os.ReadDir(".")
	if err == nil {
		names := map[string]bool{}
		var listed []string
		for _, e := range entries {
			names[e.Name()] = true
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			name := e.Name()
			if e.IsDir() {
				name += "/"
			}
			listed = append(listed, name)
		}

		var detected []string
		seen := map[string]bool{}
		for _, m := range projectMarkers {
			if names[m.file] && !seen[m.label] {
				detected = append(detected, m.label)
				seen[m.label] = true
			}
		}
		for name := range names {
			if strings.HasSuffix(name, ".xcodeproj") || strings.HasSuffix(name, ".xcworkspace") {
				if !seen["Xcode"] {
					detected = append(detected, "Xcode")
					seen["Xcode"] = true
				}
			}
		}
		if len(detected) > 0 {
			lines = append(lines, "Project: "+strings.Join(detected, ", "))
		}

		sort.Strings(listed)
		if len(listed) > maxListedEntries {
			extra := len(listed) - maxListedEntries
			listed = append(listed[:maxListedEntries], fmt.Sprintf("...and %d more", extra))
		}
		if len(listed) > 0 {
			lines = append(lines, "Contents: "+strings.Join(listed, ", "))
		}
	}

	return strings.Join(lines, "\n")
}
