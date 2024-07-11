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

func checkQueryType(query string) string {
	systemPrompt := "You are an input categorizer. Figure out if the user's prompt is an ERROR (question about a bug, an error message or stacktrace), SHELL_TASK (a prompt that can be answered only with shell commands), CODE_QUESTION (a question that can be answered with code), FILE_QUESTION (a question about a file in the project), PROJECT_QUESTION (a question about the project codebase), OTHER_QUESTION (another question you can answer), or UNKNOWN (you can't categorize). Do not output ANY text except the one word category."
	userPrompt := query
	return makeQuery(systemPrompt, userPrompt, false)
}

func checkIntent(query string) string {
	systemPrompt := "You are an intent classifier. Make a judgement if the output is likely to be a SNIPPET (a few lines of code or commands, copy/pasteable), CODE_UPDATES (one or more project files undergo sizable edits), ANSWER (an explaination or information), or UNKNOWN (you can't categorize). Do not output ANY text except the one word category."
	userPrompt := query
	return makeQuery(systemPrompt, userPrompt, false)
}

func checkError(errorMessage string) string {
	systemPrompt := "You are an expert at diagnosing errors. The error is provided by the user, along with potentially relevant files. Try to analyze the error and provide a solution. If you aren't confident, don't answer. If you need more information, ask for it."
	userPrompt := errorMessage
	return makeQuery(systemPrompt, userPrompt, true)
}

func askQuestion(question string, qType string) string {
	systemPrompt := "Answer the question"
	if qType == "code" {
		systemPrompt = "You are an expert at writing code from a user's instructions. Output ONLY code that can be pasted directly into a code editor. Minimize surrounding code and comments, focus only on returning the most relevant part. Don't use markdown (plaintext only)."
	} else if qType == "other" {
		systemPrompt = "You are answering a question from within a shell window. Be as concise as possible. Don't use markdown (plaintext only). If you need more information, ask for it."
	} else if qType == "project" {
		//Get from vectordb
	}
	userPrompt := question
	return makeQuery(systemPrompt, userPrompt, true)
}

func createShellCommand(task string) string {
	systemPrompt := "You are an expert at writing shell commands from a user's instructions. Output ONLY line separated commands that can be pasted directly into a terminal."
	userPrompt := task
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
