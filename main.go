package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	lop "github.com/samber/lo/parallel"
	openai "github.com/sashabaranov/go-openai"
)

type Page struct {
	Text                    string `json:"text"`
	IllustrationDescription string `json:"alt"`
	Illustration            string `json:"imageUrl"`
	Voiceover               string `json:"audioUrl"`
}

type MainStoryResult struct {
	pages []Page
}

func createNewStory(client *openai.Client, storyNum int, topic string) error {

	if err := os.Mkdir(fmt.Sprintf("story%v", storyNum), 0666); err != nil {
		slog.Error("Error creating directory: %v\n", err)
		return err
	}

	slog.Debug("Writing story...")
	story, _ := createStory(client, topic)

	slog.Debug("Splitting story into pages...")
	pagesAndIllustraitons, _ := splitIntoPages(client, story)
	slog.Debug(fmt.Sprintf("pages and illustrations: %v", pagesAndIllustraitons))

	texts := lop.Map(pagesAndIllustraitons.Pages, func(page PageResult, _ int) string {
		return page.Text
	})
	texts[len(texts)-1] += " The End."

	pages := texts
	imagePrompts := lop.Map(pagesAndIllustraitons.Pages, func(page PageResult, _ int) string {
		response, _ := createIllustrationDescription(client, page.IllustrationDescription, pagesAndIllustraitons.IllustrationStyle, pagesAndIllustraitons.CharacterDescriptions)
		return response
	})

	slog.Debug("Creating illustrations...")
	illustrations := lop.Map(imagePrompts, func(prompt string, _ int) []byte {
		illustration, _ := createIllustration(prompt)
		return illustration
	})

	var result MainStoryResult
	result.pages = make([]Page, 0)

	var illustration []byte
	for i, page := range pages {
		illustration = illustrations[i]
		slog.Debug("Writing illustration to file...")
		if err := os.WriteFile(fmt.Sprintf("story%v/page%v.png", storyNum, i+1), illustration, 0666); err != nil {
			slog.Error("Error writing illustration to file: %v\n", err)
		}
		result.pages = append(result.pages, Page{
			Text:                    page,
			IllustrationDescription: imagePrompts[i],
			Illustration:            fmt.Sprintf("story%v/page%v.png", storyNum, i+1),
			Voiceover:               "",
		})
	}

	slog.Debug("Creating voiceovers...")
	voiceovers, err := createVoiceovers(client, pages)
	if err != nil {
		slog.Error("Error creating voiceovers: %v\n", err)
		return err
	}
	slog.Debug("Voiceovers created.")

	for i, voiceover := range voiceovers {
		slog.Debug("Writing voiceover to file...")
		if err := os.WriteFile(fmt.Sprintf("story%v/page%v.mp3", storyNum, i+1), voiceover, 0666); err != nil {
			slog.Error("Error writing voiceover to file: %v\n", err)
		}
		result.pages[i].Voiceover = fmt.Sprintf("story%v/page%v.mp3", storyNum, i+1)
	}

	for i, page := range result.pages {
		page.IllustrationDescription = imagePrompts[i]
	}

	json_story, err := json.Marshal(result.pages)
	if err != nil {
		slog.Error("Error marshalling story: %v\n", err)
		return err
	}

	if err := os.WriteFile(fmt.Sprintf("story%v/story%v.json", storyNum, storyNum), json_story, 0666); err != nil {
		slog.Error("Error writing story to file: %v\n", err)
	}

	return nil
}

func main() {
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))

	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file")
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	err = createNewStory(client, 21, "")
	if err != nil {
		slog.Error("Error creating story: %v\n", err)
		return
	}
}
