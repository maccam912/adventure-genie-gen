package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/imroc/req/v3"
)

var URL = "https://api.stability.ai/v1/generation/stable-diffusion-xl-1024-v1-0/text-to-image"

type Prompt struct {
	Text   string `json:"text"`
	Weight int    `json:"weight"`
}

type StabilityBody struct {
	Steps       int      `json:"steps"`
	Width       int      `json:"width"`
	Height      int      `json:"height"`
	Seed        int      `json:"seed"`
	CfgScale    int      `json:"cfg_scale"`
	Samples     int      `json:"samples"`
	StylePreset string   `json:"style_preset"`
	TextPrompts []Prompt `json:"text_prompts"`
}

type Artifact struct {
	Base64       string `json:"base64"`
	FinishReason string `json:"finishReason"`
	Seed         int    `json:"seed"`
}

type Result struct {
	Artifacts []Artifact `json:"artifacts"`
}

func DefaultStabilityBody() StabilityBody {
	posPrompt := Prompt{
		Text:   "A render of a curling stone",
		Weight: 1,
	}

	negPrompt := Prompt{
		Text:   "blurry, bad",
		Weight: -1,
	}

	return StabilityBody{
		Steps:       40,
		Width:       1024,
		Height:      1024,
		Seed:        0,
		CfgScale:    5,
		Samples:     1,
		StylePreset: "fantasy-art",
		TextPrompts: []Prompt{posPrompt, negPrompt},
	}
}

func createIllustration(prompt string) ([]byte, error) {
	posPrompt := Prompt{
		Text:   prompt,
		Weight: 1,
	}

	request := DefaultStabilityBody()
	request.TextPrompts[0] = posPrompt

	client := req.C().DevMode()

	var result Result

	resp, err := client.R().
		SetBody(request).
		SetSuccessResult(&result).
		SetHeader("Authorization", "Bearer "+os.Getenv("STABILITY_API_KEY")).
		Post(URL)
	if err != nil {
		log.Fatal(err)
	}

	if !resp.IsSuccessState() {
		fmt.Println("bad response status:", resp.Status)
		return []byte{}, err
	}

	png_bytes, err := base64.StdEncoding.DecodeString(result.Artifacts[0].Base64)
	if err != nil {
		fmt.Println("error:", err)
		return []byte{}, err
	}

	return png_bytes, nil
}
