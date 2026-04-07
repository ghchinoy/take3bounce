package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func main() {
	godotenv.Load()
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: os.Getenv("GOOGLE_CLOUD_LOCATION"),
	})
	if err != nil {
		log.Fatal(err)
	}

	ttsResp, err := client.Models.GenerateContent(ctx, "gemini-3.1-flash-tts-preview", genai.Text("Hello world!"), &genai.GenerateContentConfig{
		ResponseModalities: []string{"AUDIO"},
		SpeechConfig: &genai.SpeechConfig{
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: "Aoede",
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(ttsResp.Candidates) > 0 && len(ttsResp.Candidates[0].Content.Parts) > 0 {
		for _, part := range ttsResp.Candidates[0].Content.Parts {
			if part.InlineData != nil {
				data := part.InlineData.Data
				fmt.Printf("MIME: %s\n", part.InlineData.MIMEType)
				fmt.Printf("Data length: %d\n", len(data))
				if len(data) > 20 {
					fmt.Printf("Start: %q\n", string(data[:20]))
				}
			}
		}
	}
}
