package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	lop "github.com/samber/lo/parallel"

	openai "github.com/sashabaranov/go-openai"
)

type IllustrationDescription struct {
	IllustrationDescription string `json:"illustration_description"`
}

type IllustrationDescriptions struct {
	IllustrationDescriptions []string `json:"illustration_descriptions"`
}

func getCompletion(client *openai.Client, system_message string, user_message []string, temperature float32) (string, error) {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: system_message,
		},
	}

	for _, message := range user_message {
		messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: message})
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:          openai.GPT4TurboPreview,
			Messages:       messages,
			Temperature:    temperature,
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func createVoiceover(client *openai.Client, prompt string) ([]byte, error) {
	status := 429
	var resp io.ReadCloser
	var err error
	for status == 429 {
		resp, err = client.CreateSpeech(context.Background(), openai.CreateSpeechRequest{
			Input:          prompt,
			Model:          openai.TTSModel1,
			Voice:          openai.VoiceNova,
			ResponseFormat: openai.SpeechResponseFormatMp3,
		})

		if err == nil {
			status = 200
		}

		e := &openai.APIError{}
		if errors.As(err, &e) {
			if e.HTTPStatusCode == 429 || e.HTTPStatusCode == 500 {
				status = 429
				time.Sleep(1 * time.Minute)
			} else {
				status = 400
			}
		} else {
			status = 200
		}
	}

	if err != nil {
		fmt.Printf("CreateSpeech error: %v\n", err)
		return []byte{}, err
	}

	mp3_bytes, err := io.ReadAll(resp)
	if err != nil {
		fmt.Println("error:", err)
		return []byte{}, err
	}
	return mp3_bytes, nil
}

func createVoiceovers(client *openai.Client, pages []string) ([][]byte, error) {
	voiceovers := lop.Map(pages, func(page string, _ int) []byte {
		resp, _ := createVoiceover(client, page)
		return resp
	})

	return voiceovers, nil
}
