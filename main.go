package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
)

const (
	apiEndpoint = "https://api.openai.com/v1/chat/completions"
)

func main() {
	for {
		// read input from terminal
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter prompt: ")
		text, _ := reader.ReadString('\n')
		// send text to GPT API
		gptResponse, err := sendToGPT(text)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		// print generated text
		fmt.Println("Generated text:", gptResponse)
	}
}

func sendToGPT(inputText string) (string, error) {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Error while opening config file: %v", err)
	}
	defer file.Close()

	// Decode JSON from the file into the Config struct
	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error while decoding config file: %v", err)

	}

	// Access the API key
	apiKey := config.APIKey
	client := resty.New()

	response, err := client.R().
		SetAuthToken(apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"model":      "gpt-3.5-turbo",
			"messages":   []interface{}{map[string]interface{}{"role": "system", "content": inputText}},
			"max_tokens": 50,
		}).
		Post(apiEndpoint)

	if err != nil {
		log.Fatalf("Error while sending send the request: %v", err)
	}
	if response.StatusCode() != 200 {
		return "", fmt.Errorf("%s", response.Body())
	}
	body := response.Body()

	var chatCompletion ChatCompletion
	err = json.Unmarshal([]byte(body), &chatCompletion)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return "", fmt.Errorf("%s", err)
	}
	contentValue := chatCompletion.Choices[0].Message.Content

	return contentValue, nil
}

type ChatCompletion struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type Config struct {
	APIKey string `json:"apiKey"`
}
