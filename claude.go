package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	Messages    []Message `json:"messages"`
	System      string    `json:"system,omitempty"`
	Stream      bool      `json:"stream"`
}

type AnthropicStreamResponse struct {
	Type  string `json:"type"`
	Index int    `json:"index,omitempty"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta,omitempty"`
}

type Config struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Url         string  `json:"url"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	TopP        float64 `json:"top_p"`
}

// func determineClarity(task string) string{
// 	systemPrompt := "You screen the user's question for ambiguity. If could somewhat likely answer the question without more information, output 'Clear'. If you need more information, output a specific follow up question. Provide options if there are 3 or fewer scenarios with a 95% likelihood. Don't answer the question."
// 	userPrompt := "Output a follow up question if needed. Otherwise, output 'Clear'. Do not answer the question. Input: " + task
// 	answer := makeQuery(systemPrompt, userPrompt, false, nil)
// 	if strings.TrimSpace(strings.ToLower(answer)) == "clear" {
// 		return ""
// 	} else {
// 		return answer
// 	}
// }

// pls cmd
func createShellCommand(task string, lsOutput string, pwdOutput string, currentBranch string, print bool, history *[]string) (string, bool) {
	// followUp := ""
	// if history == nil {
	// 	followUp = determineClarity(task)
	// }
	// if followUp == "" {
	systemPrompt := "You are an expert at writing macos compatible shell commands from a user's instructions. Output ONLY line separated commands that can be pasted directly into a terminal. Commands must be executed directly from the terminal without opening any interactive interfaces, allowing everything to be executed seamlessly in a single step."
	userPrompt := task
	return makeQuery(systemPrompt, userPrompt, print, history), true
	// } else {
	// 	fmt.Println("")
	// 	fmt.Println(followUp)
	// 	fmt.Println("")
	// 	return followUp, false
	// }
}

// pls write
func answerQuestion(question string) (string, bool) {
	// followUp := determineClarity(question)
	// if followUp == "" {
	systemPrompt := "Answer the question as concisely as possible. You can use markdown formatting including headers, lists, code blocks, and emphasis to make your response clear and well-structured."
	userPrompt := question
	return makeQuery(systemPrompt, userPrompt, true, nil), true
	// } else {
	// 	fmt.Println("")
	// 	fmt.Println(followUp)
	// 	fmt.Println("")
	// 	return followUp, false
	// }
}

// pls explain
func explainEachLine(content string, prompt string) string {
	systemPrompt := "You are an expert at explaining shell commands, code, regex, and other programming syntax. You can use markdown formatting to make your explanations clear and well-structured. "
	if prompt == "" {
		systemPrompt += "The user will provide an input, explain what each line does in no more than 1 sentence. Output {line}: {your explaination} for each line. Do not output any other text."
	} else {
		systemPrompt += "The user will provide an input. Concisely answer the following about it: " + prompt
	}
	userPrompt := content
	return makeQuery(systemPrompt, userPrompt, true, nil)
}

func followUp(input string, action string, output string, userPrompt string) string {
	systemPrompt := ""
	if action == "sh" {
		systemPrompt = "You are an expert helping the user with the macos shell."
	} else if action == "write" {
		systemPrompt = "You are an expert at writing regex, code, and other programming syntax."
	} else if action == "explain" {
		systemPrompt = "You are an interface within a macos terminal shell. Your job is to explain shell commands, code, regex, and answer other programming related questions."
	} else if action == "check" {
		systemPrompt = "You are an expert at analyzing git diffs for issues."
	}

	systemPrompt += "Answer the user's question clearly but as concisely as possible. You can use markdown formatting to make your response clear and well-structured."

	history := []string{input, action, output}

	return makeQuery(systemPrompt, userPrompt, true, &history)
}

// pls check
func analyzeDiff(diff string) string {
	systemPrompt := "You are an expert at analyzing git diffs for issues. Output any issues found in the diff. If no issues are found, output 'No issues found'. You can use markdown formatting to make your response clear and well-structured. Only raise glaring issues that could cause real problems, not any nitpicky issues. Show lines prepended with + as additions and lines prepended with - as deletions."
	userPrompt := diff
	return makeQuery(systemPrompt, userPrompt, true, nil)
}

// pls commit
func generateCommitMessage(diff string) string {
	systemPrompt := "You are an expert at writing concise, descriptive git commit messages. Based on the git diff provided, generate a single line commit message that follows conventional commit format when appropriate (e.g., feat:, fix:, docs:, etc.). The message should be clear, concise, and describe what was changed. Be specific! Output ONLY the commit message, no additional text or explanation."
	userPrompt := diff
	return makeQuery(systemPrompt, userPrompt, false, nil)
}

func defaultConfig() Config {
	return Config{
		Model:       "claude-sonnet-4-20250514",
		Prompt:      "",
		Url:         "https://api.anthropic.com/v1/messages",
		Temperature: 0.7,
		MaxTokens:   1000,
		TopP:        0.9,
	}
}

func getConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return defaultConfig()
	}

	configPath := filepath.Join(homeDir, ".pls_config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return defaultConfig()
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return defaultConfig()
	}

	return config
}

func setConfig(config Config) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}

	configPath := filepath.Join(homeDir, ".pls_config")

	// Convert Config struct to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		fmt.Println("Error marshalling config:", err)
		return
	}

	// Write JSON to file
	err = os.WriteFile(configPath, configJSON, 0644)
	if err != nil {
		fmt.Println("Error writing config file:", err)
	}
}

func makeQuery(systemPrompt string, userPrompt string, print bool, history *[]string) string {
	messages := []Message{}

	if history != nil {
		historyMessages := *history
		for index, message := range historyMessages {
			role := "user"
			if index%2 == 1 {
				role = "assistant"
			}
			messages = append(messages, Message{
				Role:    role,
				Content: message,
			})
		}
	}

	messages = append(messages, Message{
		Role:    "user",
		Content: userPrompt,
	})

	config := getConfig()

	requestBody := AnthropicRequest{
		Model:       config.Model,
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
		Messages:    messages,
		System:      systemPrompt,
		Stream:      true,
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error marshalling request body:", err)
		return ""
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey())
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return ""
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != 200 {
		// Read the error response
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(errorBody.Bytes(), &errorResponse); err == nil {
			if errorData, ok := errorResponse["error"].(map[string]interface{}); ok {
				if errorType, ok := errorData["type"].(string); ok {
					if message, ok := errorData["message"].(string); ok {
						switch errorType {
						case "insufficient_credit_error":
							fmt.Printf("❌ Insufficient credits: %s\n", message)
							fmt.Println("💳 Please add credits to your Anthropic account at https://console.anthropic.com/settings/billing")
						case "rate_limit_error":
							fmt.Printf("⏰ Rate limit exceeded: %s\n", message)
							fmt.Println("Please wait a moment and try again.")
						case "authentication_error":
							fmt.Printf("🔑 Authentication failed: %s\n", message)
							fmt.Println("Please run 'pls login' to set your API key.")
						case "permission_error":
							fmt.Printf("🚫 Permission denied: %s\n", message)
						case "overloaded_error":
							fmt.Printf("🔥 Service overloaded: %s\n", message)
							fmt.Println("Please try again in a few moments.")
						default:
							fmt.Printf("❌ API Error (%s): %s\n", errorType, message)
						}
					} else {
						fmt.Printf("❌ API Error: %s\n", errorType)
					}
				} else {
					fmt.Printf("❌ HTTP %d: %s\n", resp.StatusCode, errorBody.String())
				}
			} else {
				fmt.Printf("❌ HTTP %d: %s\n", resp.StatusCode, errorBody.String())
			}
		} else {
			fmt.Printf("❌ HTTP %d: %s\n", resp.StatusCode, errorBody.String())
		}
		return ""
	}

	var fullText string
	var spinner = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIndex := 0
	chunkCount := 0
	
	if print {
		fmt.Print("Thinking")
	}
	
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		
		// Handle Anthropic streaming format
		if strings.HasPrefix(line, "event: ") || strings.HasPrefix(line, "data: ") {
			if strings.HasPrefix(line, "data: ") {
				data := line[6:]
				if data == "[DONE]" {
					break
				}

				var chunk map[string]interface{}
				if err := json.Unmarshal([]byte(data), &chunk); err == nil {
					if chunkType, ok := chunk["type"].(string); ok {
						if chunkType == "error" {
							// Handle streaming errors
							if errorData, ok := chunk["error"].(map[string]interface{}); ok {
								if errorType, ok := errorData["type"].(string); ok {
									if message, ok := errorData["message"].(string); ok {
										switch errorType {
										case "overloaded_error":
											fmt.Printf("\n🔥 Service overloaded: %s\n", message)
											fmt.Println("Please try again in a few moments.")
										case "rate_limit_error":
											fmt.Printf("\n⏰ Rate limit exceeded: %s\n", message)
											fmt.Println("Please wait a moment and try again.")
										default:
											fmt.Printf("\n❌ Stream Error (%s): %s\n", errorType, message)
										}
									} else {
										fmt.Printf("\n❌ Stream Error: %s\n", errorType)
									}
								}
							}
							return ""
						} else if chunkType == "content_block_delta" {
							if delta, ok := chunk["delta"].(map[string]interface{}); ok {
								if deltaType, ok := delta["type"].(string); ok && deltaType == "text_delta" {
									if text, ok := delta["text"].(string); ok {
										fullText += text
										chunkCount++
										
										// Show loading indicator every few chunks  
										if print && chunkCount%2 == 0 {
											fmt.Printf("\r%s %s", spinner[spinnerIndex], "Thinking")
											spinnerIndex = (spinnerIndex + 1) % len(spinner)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	// Clear the loading indicator
	if print {
		fmt.Print("\r\033[K") // Clear the line
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading response body:", err)
	}

	// If we're printing, render the markdown properly
	if print && fullText != "" {
		rendered, err := renderMarkdown(fullText)
		if err != nil {
			// Fallback to simple text wrapping if glamour fails
			terminalWidth := getTerminalWidth() - 4
			fmt.Print(wrapText(fullText, terminalWidth))
		} else {
			fmt.Print(rendered)
		}
		fmt.Println("")
	}

	return fullText
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
	width, _, err := term.GetSize(0)
	if err != nil {
		width = 80
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
