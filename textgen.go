package main

import (
	"encoding/json"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

type StoryResult struct {
	Story string `json:"story"`
}

func createStory(client *openai.Client, topic string) (string, error) {
	story, err := getCompletion(client, "Write an original story appropriate for a child. It should be about five minutes long. Be creative and not too cliche. If the user has any specific requests, try to incorporate them. Respond in JSON of the form {\"story\": \"<the story you write>\"}. Characrters in the story should be animals, not humans, unless explicitly mentioned as such by the user.", []string{topic}, 1.0)

	if err != nil {
		log.Fatal(err)
	}

	var result StoryResult

	if err := json.Unmarshal([]byte(story), &result); err != nil {
		log.Fatal(err)
	}

	return result.Story, nil
}

type CharactersResult struct {
	Characters []string `json:"characters"`
}

type PageResult struct {
	Text                    string `json:"text"`
	IllustrationDescription string `json:"image"`
}

type SplitResult struct {
	Pages                 []PageResult      `json:"pages"`
	IllustrationStyle     string            `json:"illustration_style"`
	CharacterDescriptions map[string]string `json:"character_descriptions"`
}

func splitIntoPages(client *openai.Client, story string) (SplitResult, error) {
	pages, err := getCompletion(client, "Given the story provided by the user, split the text up into small pieces that will each be printed in a childrens book. Along with the text include a detailed description of the illustration that accompanies the text. Always describe the activites taking place in the image and objects to include. Pick a style of art the illustrations will use. Return the response in a JSON format: {\"pages\": [{\"text\": \"<text on the page>\", \"image\": \"<description of image>\"}], \"illustration_style\": \"<illustration style>\", \"character_descriptions\": {\"<character name>\": \"<character description>\", ...}}. ", []string{story}, 0.1)

	if err != nil {
		log.Fatal(err)
	}

	var result SplitResult

	if err := json.Unmarshal([]byte(pages), &result); err != nil {
		log.Fatal(err)
	}

	return result, nil
}

type IllustrationDescriptionResult struct {
	Description string `json:"description"`
}

func createIllustrationDescription(client *openai.Client, imageDescription string, style string, characterDescriptions map[string]string) (string, error) {
	characterDescriptionsString, err := json.Marshal(characterDescriptions)
	if err != nil {
		log.Fatal(err)
	}
	description, err := getCompletion(client, "Given the style of illustration, character descriptions, and description of the image, write a stable diffusion prompt for the image in a childrens book. Respond in JSON of the form {\"description\": \"<the description you write>\"}.", []string{"Illustration style: " + style, "Character descriptions: " + string(characterDescriptionsString), "Image description: " + imageDescription}, 0.1)
	if err != nil {
		log.Fatal(err)
	}
	return description, nil
}
