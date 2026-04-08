package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"
)

const DirectorPromptTemplate = `You are a VO director. Create a "Three-Up" based on the following text.
The text must remain 100%% IDENTICAL in all three takes, but you will insert inline audio tags to represent the emotion, pacing, and technical energy.

CRITICAL: Audio tags are NOT like XML tags. Do NOT use closing tags (e.g., never use [/happy] or [/whispering]). You must insert a single tag immediately prior to the part of the phrase that requires affective speech (e.g., "I am [laughing] very glad to see you. [short pause] Welcome!").

You may use tags from the following categories to shape the performance:

1. Non-Speech Sounds: [sigh], [laughing], [uhm]
2. Style Modifiers: [sarcasm], [robotic], [shouting], [whispering], [extremely fast]
3. Pacing and Pauses: [short pause], [medium pause], [long pause]
4. Emotional Tones (A selection):
   - Positive/High Energy: [enthusiasm], [confidence], [happy], [excitement], [joy], [triumph], [amusement]
   - Negative/Low Energy: [despair], [sadness], [boredom], [exhaustion], [melancholy], [pessimism]
   - Neutral/Analytical: [neutral], [information], [contemplative], [thoughtful], [explanation]
   - Complex/Relational: [sarcasm], [playful], [sympathy], [arrogance], [pleading], [skepticism]

Determine three pillars for each take:
- Persona: Who is speaking? (e.g., The Peer, The Sage, The Catalyst)
- Subtext: What are they thinking? (e.g., The Shrug, The Smile, The Aha)
- Technical Energy: How are they speaking? (e.g., Fast, High Pitch, Firm)

Then apply the "1, 2, 3" energy ramp:
- Take 1 (Safe): The "Safe" read. Follows punctuation closely, uses safe/neutral emotional tags and standard pacing.
- Take 2 (Pushed): The "Pushed" read. More energy, more personality. Utilizes pauses, sighs, or higher-energy emotional tags.
- Take 3 (Wildcard): The "Wildcard" read. Very casual, extreme, or unexpected. Use dramatic tags (e.g., [whispering], [laughing], [sarcasm], [despair]).

Format the output strictly as a JSON array of objects. Each object MUST have exactly these five keys:
- "take" (string, e.g. "1 - Safe")
- "persona" (string)
- "subtext" (string)
- "technicalEnergy" (string)
- "text" (string, the tagged text)

Original Text:
%s`

func handleVariations(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Ensure .env variables are available, or read from system env
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	bucketName := os.Getenv("GENMEDIA_BUCKET")
	if bucketName == "" {
		bucketName = "aaie-speech-arena"
	}
	if strings.HasPrefix(bucketName, "gs://") {
		bucketName = strings.TrimPrefix(bucketName, "gs://")
	}
	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" {
		geminiModel = "gemini-3-flash-preview"
	}
	ttsModel := os.Getenv("GEMINI_TTS_MODEL")
	if ttsModel == "" {
		ttsModel = "gemini-3.1-flash-tts-preview"
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  project,
		Location: location,
	})
	if err != nil {
		http.Error(w, "Failed to create GenAI client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	prompt := fmt.Sprintf(DirectorPromptTemplate, req.Text)

	resp, err := client.Models.GenerateContent(ctx, geminiModel,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			MaxOutputTokens:  8192,
		},
	)
	if err != nil {
		slog.Error("Failed to generate text variations", "error", err)
		http.Error(w, "Failed to generate text variations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var variations []Variation
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		var fullText string
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				fullText += part.Text
			}
		}
		if fullText != "" {
			fullText = strings.TrimSpace(fullText)
			if strings.HasPrefix(fullText, "```json") {
				fullText = strings.TrimPrefix(fullText, "```json")
				fullText = strings.TrimSuffix(fullText, "```")
				fullText = strings.TrimSpace(fullText)
			} else if strings.HasPrefix(fullText, "```") {
				fullText = strings.TrimPrefix(fullText, "```")
				fullText = strings.TrimSuffix(fullText, "```")
				fullText = strings.TrimSpace(fullText)
			}
			if err := json.Unmarshal([]byte(fullText), &variations); err != nil {
				slog.Error("JSON parse error", "error", err, "text", fullText)
				http.Error(w, "Failed to parse variations JSON", http.StatusInternalServerError)
				return
			}
		} else {
			slog.Error("Empty text variations generated")
			http.Error(w, "Empty text variations generated", http.StatusInternalServerError)
			return
		}
	} else {
		slog.Error("No variations generated")
		http.Error(w, "No variations generated", http.StatusInternalServerError)
		return
	}

	if len(variations) > 3 {
		variations = variations[:3]
	}

	var wg sync.WaitGroup
	for i := range variations {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			v := variations[idx]
			ttsPrompt := fmt.Sprintf(`# AUDIO PROFILE: Aoede
## THE SCENE: A professional voiceover studio

### DIRECTOR'S NOTES
Persona: %s
Subtext: %s
Technical: %s

#### TRANSCRIPT
%s`, v.Persona, v.Subtext, v.TechnicalEnergy, v.Text)

			var ttsResp *genai.GenerateContentResponse
			var err error
			maxRetries := 3
			for attempt := 1; attempt <= maxRetries; attempt++ {
				ttsResp, err = client.Models.GenerateContent(ctx, ttsModel,
					genai.Text(ttsPrompt), &genai.GenerateContentConfig{
						ResponseModalities: []string{"AUDIO"},
						SpeechConfig: &genai.SpeechConfig{
							VoiceConfig: &genai.VoiceConfig{
								PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
									VoiceName: "Aoede",
								},
							},
						},
					})

				if err == nil && len(ttsResp.Candidates) > 0 && len(ttsResp.Candidates[0].Content.Parts) > 0 {
					break // Success
				}
				
				var blockReason string
				if ttsResp != nil && ttsResp.PromptFeedback != nil {
					blockReason = string(ttsResp.PromptFeedback.BlockReason)
				}
				
				slog.Warn("TTS generation failed or returned empty", "variation", idx, "attempt", attempt, "error", err, "blockReason", blockReason)
				if attempt < maxRetries {
					time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
				}
			}

			if err != nil {
				slog.Error("TTS error for variation after retries", "index", idx, "error", err)
				return
			}

			if len(ttsResp.Candidates) > 0 && len(ttsResp.Candidates[0].Content.Parts) > 0 {
				for _, part := range ttsResp.Candidates[0].Content.Parts {
					if part.InlineData != nil {
						audioData := part.InlineData.Data
						mimeType := part.InlineData.MIMEType

						// Wrap raw PCM in a WAV header so browsers can play it
						if strings.HasPrefix(mimeType, "audio/l16") {
							// Gemini TTS returns 24kHz, mono, 16-bit PCM
							audioData = addWavHeader(audioData, 24000, 1, 16)
							mimeType = "audio/wav"
						}

						filename := generateFilename(mimeType)
						url, uploadErr := uploadToGCS(ctx, bucketName, filename, mimeType, audioData)
						if uploadErr != nil {
							slog.Error("Failed to upload audio to GCS", "error", uploadErr)
							return
						}
						variations[idx].Audio = url
						variations[idx].MimeType = mimeType
						break
					}
				}
			}
		}(i)
	}
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(variations)
}

func handleRetryAudio(w http.ResponseWriter, r *http.Request) {
	var req RetryAudioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	v := req.Variation

	slog.Info("Retrying/Regenerating audio for variation", "take", v.Take, "text_length", len(v.Text))

	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	bucketName := os.Getenv("GENMEDIA_BUCKET")
	if bucketName == "" {
		bucketName = "aaie-speech-arena"
	}
	if strings.HasPrefix(bucketName, "gs://") {
		bucketName = strings.TrimPrefix(bucketName, "gs://")
	}
	ttsModel := os.Getenv("GEMINI_TTS_MODEL")
	if ttsModel == "" {
		ttsModel = "gemini-3.1-flash-tts-preview"
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  project,
		Location: location,
	})
	if err != nil {
		http.Error(w, "Failed to create GenAI client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ttsPrompt := fmt.Sprintf(`# AUDIO PROFILE: Aoede
## THE SCENE: A professional voiceover studio

### DIRECTOR'S NOTES
Persona: %s
Subtext: %s
Technical: %s

#### TRANSCRIPT
%s`, v.Persona, v.Subtext, v.TechnicalEnergy, v.Text)

	var ttsResp *genai.GenerateContentResponse
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ttsResp, err = client.Models.GenerateContent(ctx, ttsModel,
			genai.Text(ttsPrompt), &genai.GenerateContentConfig{
				ResponseModalities: []string{"AUDIO"},
				SpeechConfig: &genai.SpeechConfig{
					VoiceConfig: &genai.VoiceConfig{
						PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
							VoiceName: "Aoede",
						},
					},
				},
			})

		if err == nil && len(ttsResp.Candidates) > 0 && len(ttsResp.Candidates[0].Content.Parts) > 0 {
			break
		}

		var blockReason string
		if ttsResp != nil && ttsResp.PromptFeedback != nil {
			blockReason = string(ttsResp.PromptFeedback.BlockReason)
		}

		slog.Warn("TTS retry failed or empty", "attempt", attempt, "error", err, "blockReason", blockReason)
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	if err != nil {
		slog.Error("TTS error for retry", "error", err)
		http.Error(w, "TTS failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(ttsResp.Candidates) > 0 && len(ttsResp.Candidates[0].Content.Parts) > 0 {
		for _, part := range ttsResp.Candidates[0].Content.Parts {
			if part.InlineData != nil {
				audioData := part.InlineData.Data
				mimeType := part.InlineData.MIMEType

				if strings.HasPrefix(mimeType, "audio/l16") {
					audioData = addWavHeader(audioData, 24000, 1, 16)
					mimeType = "audio/wav"
				}

				filename := generateFilename(mimeType)
				url, uploadErr := uploadToGCS(ctx, bucketName, filename, mimeType, audioData)
				if uploadErr != nil {
					slog.Error("Failed to upload audio to GCS", "error", uploadErr)
					http.Error(w, "Upload failed: "+uploadErr.Error(), http.StatusInternalServerError)
					return
				}
				v.Audio = url
				v.MimeType = mimeType
				break
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
