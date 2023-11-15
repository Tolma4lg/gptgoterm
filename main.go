package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"io"

	"github.com/go-resty/resty/v2"
	"gopkg.in/yaml.v2"
)

const (
	apiEndpoint = "https://api.openai.com/v1/chat/completions"
	modelName   = "gpt-3.5-turbo"
)

func main() {
	for {
		// read input from terminal
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter prompt: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
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

	config, err := readConfig("config.yaml")
	if err != nil {
		return "", fmt.Errorf("error while reading config file: %v", err)
	}
	apiKey := config.APIKey
	client := resty.New()

	requestBody := struct {
		Model     string        `json:"model"`
		Messages  []interface{} `json:"messages"`
		MaxTokens int           `json:"max_tokens"`
	}{modelName, []interface{}{map[string]interface{}{"role": "system", "content": inputText}}, 50}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error while encoding request body: %v", err)
	}

	response, err := client.R().
		SetAuthToken(apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBodyJSON).
		Post(apiEndpoint)

	if err != nil {
		return "", fmt.Errorf("error while sending the request: %v", err)
	}
	if response.IsError() {
		return "", fmt.Errorf("%s", response.String())
	}
	body := response.String()

	var chatCompletion ChatCompletion
	err = json.Unmarshal([]byte(body), &chatCompletion)
	if err != nil {
		return "", fmt.Errorf("error parsing JSON: %v", err)
	}
	contentValue := chatCompletion.Choices[0].Message.Content

	return contentValue, nil
}

func readConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(io.Reader(file))
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
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
	APIKey string `yaml:"apiKey"`
}
