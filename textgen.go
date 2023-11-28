package main

import (
	"encoding/json"
	"fmt"
	"log"

	lop "github.com/samber/lo/parallel"
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

func createCharacters(client *openai.Client, story string) ([]string, error) {
	characters, err := getCompletion(client, "Given the story provided by the user, generate a JSON structure of character descriptions in the form {\"characters\": [\"character 1 description\", \"character 2 description\"]}. The descriptions must be specific enough that independent image generations prompted based on the character description should produce the same visual representation of the character.", []string{story}, 0.0)

	if err != nil {
		log.Fatal(err)
	}

	var result CharactersResult

	if err := json.Unmarshal([]byte(characters), &result); err != nil {
		log.Fatal(err)
	}

	return result.Characters, nil
}

type PageResult struct {
	Text                    string `json:"text"`
	IllustrationDescription string `json:"image"`
}

type SplitResult struct {
	Pages []PageResult `json:"pages"`
}

func splitIntoPages(client *openai.Client, story string) ([]PageResult, error) {
	// pages, err := getCompletion(client, "Given the story provided by the user, generate a JSON structure of pages with their text and an illustration description in the form {\"pages\": [{\"text\": \"<text of the page>\", \"image\": \"<description of the illustration to accompany the page>\"}, {\"text\": \"<text of the page>\", \"image\": \"<description of the illustration to accompany the page>\"}, ...]}. The user will provide the text of the book you need to split. Don't say \"Page 1: ...\" as part of the text, the exact words you write will be what is printed in the book. Don't rewrite much, but correct any obvious errors in rewriting. Include a caption for the image describing what an artist should draw, ignoring character names and only describing what type of animals or people are in the image, what they are wearing or what color fur they might have for example, what is happening in the scene, etc. For example, never say \"the butterfly\" in the image description, instead say \"a butterfly with blue wings and sparlkly antennae\". Don't forget to always include an art style, which must remain consistent throughout the images, such as \"A photorealistic render\" or \"A vibrant oil painting\". For example, if the first page prompt is described as a photorealistic rendering, the prompt for the rest of the pages must also be described as photorealistic renderings. The final output should have the entire story split into a list of pages each containing a few sentences at most, along with a prompt for an image that will accompany that page in the story.", []string{story}, 0.1)
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

func createIllustrationDescription(client *openai.Client, story string, characters []string, page string) (string, error) {
	charactersStr := ""
	for _, character := range characters {
		charactersStr += character + ", "
	}
	description, err := getCompletion(client, fmt.Sprintf("You write stable diffusion prompts to illustrate pages in a book. The characters are: %v. The story you are generating the image for is: %v. For this specific response, generate a prompt to create an illustration that should accompany the text: %v. A good prompt reads like an eventual caption for the image, ignoring names but including art style or illustration style, specifying animals and fur color or other distinguishing features, any emotion the animal or subject might be displaying, and placement of objects in the scene. Someone who is blind but listens to the prompt should be able to imagine a nearly identical image. Respond in JSON in the form {\"description\": \"<image prompt>\"}.", charactersStr, story, page), []string{}, 0.5)

	if err != nil {
		log.Fatal(err)
	}

	var result IllustrationDescriptionResult
	if err := json.Unmarshal([]byte(description), &result); err != nil {
		log.Fatal(err)
	}

	return result.Description, nil
}

func createIllustrationDescriptions(client *openai.Client, pages []string, characters []string) ([]string, error) {
	story := ""
	for _, page := range pages {
		story += page + " "
	}

	illustrations := lop.Map(pages, func(page string, i int) string {
		resp, _ := createIllustrationDescription(client, story, characters, page)
		return resp
	})

	return illustrations, nil
}
