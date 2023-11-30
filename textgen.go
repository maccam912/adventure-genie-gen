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
	story, err := getCompletion(client, "Write an original story appropriate for a child. It should be about five minutes long. Be creative and not too cliche. If the user has any specific requests, try to incorporate them. Respond in JSON of the form {\"story\": \"<the story you write>\"}", []string{topic}, 1.0)

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
	Pages []PageResult `json:"pages"`
}

func splitIntoPages(client *openai.Client, story string) ([]PageResult, error) {
	pages, err := getCompletion(client, "Given the story provided by the user, generate a JSON structure of pages with their text and an illustration description in the form {\"pages\": [{\"text\": \"<text of the page>\", \"image\": \"<description of the illustration to accompany the page>\"}, {\"text\": \"<text of the page>\", \"image\": \"<description of the illustration to accompany the page>\"}, ...]}. The user will provide the text of the book you need to split. Don't say \"Page 1: ...\" as part of the text, the exact words you write will be what is printed in the book. Don't rewrite much, but correct any obvious errors in rewriting. Include a caption for the image describing what an artist should draw, ignoring character names and only describing what type of animals or people are in the image, what they are wearing or what color fur they might have for example, what is happening in the scene, etc. For example, never say \"the butterfly\" in the image description, instead say \"a butterfly with blue wings and sparlkly antennae\" and never say \"the map\", but say \"a map on faded parchment with scribbled writing on it\". Each prompt should be about 60 words long, should not specify what style of art the image should be. The descriptions will be given to separate artists so each one needs to contain all necessary information about what people or animals look like. The final output should have the entire story split into a list of pages each containing a few sentences at most, along with a prompt for an image that will accompany that page in the story.", []string{story}, 0.1)

	if err != nil {
		log.Fatal(err)
	}

	var result SplitResult

	if err := json.Unmarshal([]byte(pages), &result); err != nil {
		log.Fatal(err)
	}

	return result.Pages, nil
}

type IllustrationDescriptionResult struct {
	Description string `json:"description"`
}
