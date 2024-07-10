package main

import "fmt"

func testFunctions() {
	systemPrompt := "Call the functions in the following order: getDiff, getFile"
	userPrompt := "Call the functions in the following order: getDiff, getFile"
	response := sendWithFunctions(&systemPrompt, &userPrompt, nil)
	fmt.Println(response)
}
