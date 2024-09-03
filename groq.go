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
	systemPrompt := "Answer the question as concisely as possible, without markdown. If the user is asking for code, output ONLY code. Do not use markdown except for backticks and asterisks."
	userPrompt := question
	return makeQuery(systemPrompt, userPrompt, true, nil), true
	// } else {
	// 	fmt.Println("")
	// 	fmt.Println(followUp)
	// 	fmt.Println("")
	// 	return followUp, false
	// }
}

//pls explain
func explainEachLine(content string, prompt string) string {
	systemPrompt := "You are an expert at explaining shell commands, code, regex, and other programming syntax. You never use markdown other than backticks. "
	if prompt == "" {
		systemPrompt += "The user will provide an input, explain what each line does in no more than 1 sentence. Output {line}: {your explaination} for each line. Do not output any other text."
	} else {
		systemPrompt += "The user will provide an input. Concisely answer the following about it: " + prompt
	}
	userPrompt := content
	return makeQuery(systemPrompt, userPrompt, true, nil)
}

func followUp(input string, action string, output string, userPrompt string) string{
	systemPrompt := ""
	if action == "do" {
		systemPrompt = "You are an expert helping the user with the macos shell."
	} else if action == "write" {
		systemPrompt = "You are an expert at writing regex, code, and other programming syntax."
	} else if action == "explain" {
		systemPrompt = "You are an interface within a macos terminal shell. Your job is to explain shell commands, code, regex, and answer other programming related questions."
	} else if action == "check" {
		systemPrompt = "You are an expert at analyzing git diffs for issues."
	}

	systemPrompt += "Answer the user's question clearly but as concisely as possible. Do not use markdown except for backticks and asterisks."

	history := []string{input, action, output}

	return makeQuery(systemPrompt, userPrompt, true, &history)
}

//pls check
func analyzeDiff(diff string) string {
	systemPrompt := "You are an expert at analyzing git diffs for issues. Output any issues found in the diff. If no issues are found, output 'No issues found'. Do not use markdown except for backticks and asterisks."
	userPrompt := diff
	return makeQuery(systemPrompt, userPrompt, true, nil)
}

func makeQuery(systemPrompt string, userPrompt string, print bool, history *[]string) string {
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	if history != nil {
		historyMessages := *history
		for index, message := range historyMessages {
			role := "user"
			if index % 2 == 1 {
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
	req.Header.Set("Authorization", "Bearer "+apiKey())

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
	var fullText string
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
						fullText += content.(string)

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

	return strings.ReplaceAll(fullText, "`", "")
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
			if j % 2 == 1 {
				styledText += (codeStart + codeSections[j] + codeEnd)
			} else {
				styledText += codeSections[j]
			}
		}
		wrappedParagraphs[i] = styledText
	}

	return strings.Join(wrappedParagraphs, "\n")
}