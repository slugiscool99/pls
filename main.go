package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pls <command>",
	Short: "AI in your shell",
	Long:  `AI-powered command line assistant.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			printHelp()
			os.Exit(0)
		}

		action := args[0]

		if action == "login" { //login to pls
			if len(args) > 1 {
				setApiKey(args[1])
			} else {
				addApiKey()
			}
		} else if action == "help" { //print help
			printHelp()
		} else if action == "model" { //set model
			if len(args) > 1 {
				setModel(args[1])
			} else {
				fmt.Println("Current model:", getModel())
			}
		} else { //run the command
			query := strings.Join(args, " ")
			runCmd(query)
		}
	},
}

func runCmd(query string) {
	response := callGroqAPI(query)
	if response != "" {
		rendered, err := renderMarkdown(response)
		if err != nil {
			// Fallback to simple text wrapping if glamour fails
			terminalWidth := getTerminalWidth() - 4
			fmt.Print(wrapText(response, terminalWidth))
		} else {
			fmt.Print(rendered)
		}
		fmt.Println("")
	}
}

func callGroqAPI(prompt string) string {
	apiKey := apiKey()
	if apiKey == "" {
		fmt.Println("Please set your Groq API key first with: pls login <your-api-key>")
		os.Exit(1)
	}

	model := getModel()
	if model == "" {
		model = "openai/gpt-oss-20b" // default model
	}

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are an AI assistant in a command line interface. Your responses will be displayed in the terminal. Please be concise and clear in your output.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Sprintf("Error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response: %v", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Sprintf("API Error (%d): %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Sprintf("Error parsing response: %v", err)
	}

	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "No response from API"
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "Invalid response format"
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "Invalid message format"
	}

	content, ok := message["content"].(string)
	if !ok {
		return "No content in response"
	}

	return content
}

func getModel() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "openai/gpt-oss-20b"
	}

	configPath := filepath.Join(homeDir, ".pls_model")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "openai/gpt-oss-20b"
	}

	return strings.TrimSpace(string(data))
}

func setModel(model string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}

	configPath := filepath.Join(homeDir, ".pls_model")
	err = os.WriteFile(configPath, []byte(model), 0644)
	if err != nil {
		fmt.Println("Error saving model:", err)
		return
	}

	fmt.Printf("Model set to: %s\n", model)
}

func setApiKey(key string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}

	configPath := filepath.Join(homeDir, ".pls_key")
	err = os.WriteFile(configPath, []byte(key), 0600)
	if err != nil {
		fmt.Println("Error saving API key:", err)
		return
	}

	fmt.Println("API key saved successfully.")
}

func renderMarkdown(text string) (string, error) {
	// Create a glamour renderer with auto-detection of terminal theme
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(getTerminalWidth()-4), // Leave some margin
	)
	if err != nil {
		return "", err
	}

	// Render the markdown first
	rendered, err := r.Render(text)
	if err != nil {
		return "", err
	}

	// Then apply git diff coloring to the rendered output
	coloredText := applyDiffColoring(rendered)

	return coloredText, nil
}

func applyDiffColoring(text string) string {
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if strings.HasPrefix(line, "+") {
			// Green background with black text for additions
			result = append(result, "\033[42m\033[30m"+line+"\033[0m")
		} else if strings.HasPrefix(line, "-") {
			// Red background with white text for deletions
			result = append(result, "\033[41m\033[37m"+line+"\033[0m")
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func getTerminalWidth() int {
	cmd := exec.Command("tput", "cols")
	output, err := cmd.Output()
	if err != nil {
		return 80 // default width
	}
	var width int
	fmt.Sscanf(string(output), "%d", &width)
	if width <= 0 {
		return 80
	}
	return width
}

func wrapText(text string, lineWidth int) string {
	paragraphs := strings.Split(text, "\n")
	var wrappedParagraphs []string

	codeStart := "\033[1m"
	codeEnd := "\033[0m"

	for _, paragraph := range paragraphs {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			wrappedParagraphs = append(wrappedParagraphs, "")
			continue
		}

		var lines []string
		currentLine := words[0]

		for _, word := range words[1:] {
			if len(currentLine)+len(word)+1 <= lineWidth {
				currentLine += " " + word
			} else {
				lines = append(lines, currentLine)
				currentLine = word
			}
		}

		lines = append(lines, currentLine)
		wrappedParagraphs = append(wrappedParagraphs, strings.Join(lines, "\n"))
	}

	// Wrap code within backticks with ANSI bold codes
	for i := 0; i < len(wrappedParagraphs); i++ {
		styledText := ""

		allSingleBackticks := strings.ReplaceAll(wrappedParagraphs[i], "```", "`")
		codeSections := strings.Split(allSingleBackticks, "`")
		for j := 0; j < len(codeSections); j += 1 {
			if j%2 == 1 {
				styledText += (codeStart + codeSections[j] + codeEnd)
			} else {
				styledText += codeSections[j]
			}
		}
		wrappedParagraphs[i] = styledText
	}

	return strings.Join(wrappedParagraphs, "\n")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
