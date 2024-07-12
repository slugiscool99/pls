package main

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func send(systemPrompt *string, userPrompt *string, messageArray *[]openai.ChatCompletionMessage) string {
	messages := []openai.ChatCompletionMessage{}
	if messageArray != nil {
		messages = *messageArray
	} else if userPrompt != nil && systemPrompt != nil {
		messages = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: *systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: *userPrompt,
			},
		}
	} else {
		return "Couldn't find your input"
	}

	apiKey := apiKey()
	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT40613,
			Messages: messages,
		},
	)

	if err != nil {
		reportError("sendChat", err)
		return "Please try again"
	}

	return resp.Choices[0].Message.Content
}

func sendWithFunctions(systemPrompt *string, userPrompt *string, messageArray *[]openai.ChatCompletionMessage) string {
	tools := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "getDiff",
				Description: "Runs git diff on a file",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"filePath": {
							Type:        jsonschema.String,
							Description: "The absolute path to the file",
						},
					},
					Required: []string{"filePath"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "getFile",
				Description: "Returns the contents of a file",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"filePath": {
							Type:        jsonschema.String,
							Description: "The absolute path to the file",
						},
					},
					Required: []string{"filePath"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "vectorSearch",
				Description: "Query a vector search of the repo",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"filePath": {
							Type:        jsonschema.String,
							Description: "The absolute path to the file",
						},
					},
					Required: []string{"filePath"},
				},
			},
		},
	}

	var messages []openai.ChatCompletionMessage
	if messageArray != nil {
		messages = *messageArray
	} else if userPrompt != nil && systemPrompt != nil {
		messages = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: *systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: *userPrompt,
			},
		}
	} else {
		return "Couldn't find your input"
	}

	apiKey := apiKey()
	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:      openai.GPT40613,
			Messages:   messages,
			Tools:      tools,
			ToolChoice: "required",
		},
	)
	// msg := resp.Choices[0].Message
	// fmt.Println("count ", len(msg.ToolCalls))

	if err != nil {
		reportError("sendChat", err)
		return "Please try again"
	}

	return resp.Choices[0].Message.Content
}

func reportError(function string, err error) {
	fmt.Println(function+" Error:", err)
}
