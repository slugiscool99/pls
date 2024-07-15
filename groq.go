package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/term"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	TopP        float64   `json:"top_p"`
	Stream      bool      `json:"stream"`
	Stop        *string   `json:"stop"`
}

const groqKey = "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"

// pls cmd
func createShellCommand(task string, print bool) string {
	systemPrompt := "You are an expert at writing macos compatible shell commands from a user's instructions. Output ONLY line separated commands that can be pasted directly into a terminal. Commands must be executed directly from the terminal without opening any interactive interfaces, allowing everything to be executed seamlessly in a single step."
	userPrompt := task
	return makeQuery(systemPrompt, userPrompt, print)
}

func explainEachLine(content string, prompt string) string {
	systemPrompt := "You are an expert at explaining shell commands, code, regex, and other programming syntax. "
	if prompt == "" {
		systemPrompt += "The user has provided an input, explain what each line does in no more than 1 sentence. Output {line}: {your explaination} for each line. Do not output any other text."
	} else {
		systemPrompt += "The user has provided an input. Answer the following about the code: " + prompt
	}
	userPrompt := content
	return makeQuery(systemPrompt, userPrompt, true)
}

// pls code
func answerQuestion(question string, lang string) string {
	language := ""
	if lang != "" {
		language += "Unless specified otherwise, use " + lang + " syntax."
	}
	systemPrompt := "Answer the question as concisely as possible, without markdown. If the user is asking for code, output ONLY code. " + language
	userPrompt := question
	return makeQuery(systemPrompt, userPrompt, true)
}

func makeQuery(systemPrompt string, userPrompt string, print bool) string {
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userPrompt,
		},
	}

	requestBody := ChatCompletionRequest{
		Messages:    messages,
		Model:       "llama3-70b-8192",
		Temperature: 1,
		MaxTokens:   1024,
		TopP:        1,
		Stream:      true,
		Stop:        nil,
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error marshalling request body:", err)
		return ""
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+groqKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return ""
	}
	defer resp.Body.Close()

	if print {
		fmt.Println("")
	}

	var accumulatedText string
	terminalWidth := getTerminalWidth() - 40
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "data: [DONE]" {
			break
		}

		// Print each data line received
		if len(line) > 6 && line[:6] == "data: " {
			data := line[6:]
			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err == nil {
				choices := chunk["choices"].([]interface{})
				if len(choices) > 0 {
					delta := choices[0].(map[string]interface{})["delta"]
					if content, ok := delta.(map[string]interface{})["content"]; ok {
						accumulatedText += content.(string)

						// Check if we have a complete paragraph
						if strings.Contains(accumulatedText, "\n\n") {
							paragraphs := strings.Split(accumulatedText, "\n\n")
							for i := 0; i < len(paragraphs)-1; i++ {
								if print {
									fmt.Println(wrapText(paragraphs[i], terminalWidth))
									fmt.Println() // Add an empty line between paragraphs
								}
							}
							accumulatedText = paragraphs[len(paragraphs)-1]
						}
					}
				}
			}
		}
	}

	if accumulatedText != "" && print {
		fmt.Print(wrapText(accumulatedText, terminalWidth))
	}

	if print {
		fmt.Println("")
		fmt.Println("")
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading response body:", err)
	}

	return accumulatedText
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

	return strings.Join(wrappedParagraphs, "\n")
}
