// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
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


const OneUpPromptTemplate = `You are a VO director. Create a "One-Up" based on the following text and reading tone.
The core text must remain 100%% IDENTICAL, but you will insert inline audio tags to represent the emotion, pacing, and technical energy requested by the "Reading Tone" or "Vibe".

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

Based on the requested Reading Tone: "%s", determine three pillars for the take:
- Persona: Who is speaking?
- Subtext: What are they thinking?
- Technical Energy: How are they speaking?

Format the output strictly as a JSON array containing EXACTLY ONE object. The object MUST have exactly these five keys:
- "take" (string, e.g. "Custom - %s")
- "persona" (string)
- "subtext" (string)
- "technicalEnergy" (string)
- "text" (string, the tagged text)

Original Text:
%s`

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

// handleVariations is the primary HTTP endpoint for the orchestrator.
// It accepts a JSON payload containing the user's raw text and selected
// VoiceActor. It performs a two-step generation process:
// 1. Calls Gemini Pro to generate three distinct text variations (Takes).
// 2. Spawns parallel Goroutines to call Gemini Flash TTS for each variation,
//    synthesizing the audio and uploading the raw PCM byte streams to GCS.
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
			// Resiliently extract JSON array if the model wrapped it in markdown or chat text
			startIdx := strings.Index(fullText, "[")
			endIdx := strings.LastIndex(fullText, "]")
			if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
				fullText = fullText[startIdx : endIdx+1]
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
			voiceName := req.VoiceActor.BaseVoice
			if voiceName == "" {
				voiceName = "Aoede" // Fallback
			}
			stylePrompt := req.VoiceActor.StylePrompt
			if stylePrompt == "" {
				stylePrompt = "# AUDIO PROFILE: Default\n## THE SCENE: A professional studio"
			}

			ttsPrompt := fmt.Sprintf(`%s

### VARIATION SPECIFIC DIRECTION
Persona: %s
Subtext: %s
Technical: %s

#### TRANSCRIPT
%s`, stylePrompt, v.Persona, v.Subtext, v.TechnicalEnergy, v.Text)

			var ttsResp *genai.GenerateContentResponse
			var err error
			maxRetries := MaxTTSRetries
			for attempt := 1; attempt <= maxRetries; attempt++ {
				
				// Fallback to 2.5-flash-tts if 3.1 fails with PROHIBITED_CONTENT multiple times
				currentModel := ttsModel
				if attempt == maxRetries {
					currentModel = "gemini-2.5-flash-tts"
					slog.Info("Falling back to backup TTS model", "variation", idx, "model", currentModel)
				}

				slog.Info("Generating TTS", 
					"takeIndex", idx, 
					"voicePreset", req.VoiceActor.ShortName, 
					"model", currentModel, 
					"text", v.Text)
				ttsResp, err = client.Models.GenerateContent(ctx, currentModel,
					genai.Text(ttsPrompt), &genai.GenerateContentConfig{
						ResponseModalities: []string{"AUDIO"},
						SpeechConfig: &genai.SpeechConfig{
					VoiceConfig: &genai.VoiceConfig{
						PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
							VoiceName: voiceName,
						},
					},
				},
				SafetySettings: []*genai.SafetySetting{
					{
						Category:  genai.HarmCategoryHarassment,
						Threshold: genai.HarmBlockThresholdBlockNone,
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
						if strings.HasPrefix(strings.ToLower(mimeType), "audio/l16") {
							// Gemini TTS returns 24kHz, mono, 16-bit PCM
							audioData = addWavHeader(audioData, PCMDefaultSampleRate, PCMDefaultChannels, PCMDefaultBitDepth)
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

// handleRetryAudio allows the frontend to manually request a single TTS
// generation. This is typically invoked when a previous generation failed
// (e.g., due to safety filters) or when the user manually tweaks the markup
// tags in the UI and wants to hear the new result without regenerating text.
// It implements an exponential backoff retry loop and falls back to an older
// model if the primary model consistently returns PROHIBITED_CONTENT.
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

	voiceName := req.VoiceActor.BaseVoice
	if voiceName == "" {
		voiceName = "Aoede"
	}
	stylePrompt := req.VoiceActor.StylePrompt
	if stylePrompt == "" {
		stylePrompt = "# AUDIO PROFILE: Default\n## THE SCENE: A professional studio"
	}

	ttsPrompt := fmt.Sprintf(`%s

### VARIATION SPECIFIC DIRECTION
Persona: %s
Subtext: %s
Technical: %s

#### TRANSCRIPT
%s`, stylePrompt, v.Persona, v.Subtext, v.TechnicalEnergy, v.Text)

	var ttsResp *genai.GenerateContentResponse
	maxRetries := MaxTTSRetries
	for attempt := 1; attempt <= maxRetries; attempt++ {
		
		// Fallback to 2.5-flash-tts on last attempt
		currentModel := ttsModel
		if attempt == maxRetries {
			currentModel = "gemini-2.5-flash-tts"
			slog.Info("Falling back to backup TTS model", "take", v.Take, "model", currentModel)
		}

		slog.Info("Generating TTS (Retry)", 
			"take", v.Take, 
			"voicePreset", req.VoiceActor.ShortName, 
			"model", currentModel, 
			"text", v.Text)
		ttsResp, err = client.Models.GenerateContent(ctx, currentModel,
			genai.Text(ttsPrompt), &genai.GenerateContentConfig{
				ResponseModalities: []string{"AUDIO"},
				SpeechConfig: &genai.SpeechConfig{
					VoiceConfig: &genai.VoiceConfig{
						PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
							VoiceName: voiceName,
						},
					},
				},
				SafetySettings: []*genai.SafetySetting{
					{
						Category:  genai.HarmCategoryHarassment,
						Threshold: genai.HarmBlockThresholdBlockNone,
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

				if strings.HasPrefix(strings.ToLower(mimeType), "audio/l16") {
					audioData = addWavHeader(audioData, PCMDefaultSampleRate, PCMDefaultChannels, PCMDefaultBitDepth)
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

func handleGenerateOne(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req GenerateOneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

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
		geminiModel = "gemini-1.5-pro"
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

	prompt := fmt.Sprintf(OneUpPromptTemplate, req.ReadingTone, req.ReadingTone, req.Text)

	resp, err := client.Models.GenerateContent(ctx, geminiModel,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			MaxOutputTokens:  8192,
		},
	)
	if err != nil {
		slog.Error("Failed to generate text variation", "error", err)
		http.Error(w, "Failed to generate text variation: "+err.Error(), http.StatusInternalServerError)
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
		
		fullText = strings.TrimPrefix(fullText, "```json")
		fullText = strings.TrimSuffix(fullText, "```")
		fullText = strings.TrimSpace(fullText)

		if err := json.Unmarshal([]byte(fullText), &variations); err != nil {
			slog.Error("Failed to parse LLM JSON", "error", err, "raw", fullText)
			http.Error(w, "Failed to parse LLM output: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		slog.Error("No variation generated")
		http.Error(w, "No variation generated", http.StatusInternalServerError)
		return
	}

	if len(variations) == 0 {
		http.Error(w, "No variation returned from LLM", http.StatusInternalServerError)
		return
	}
	
	v := variations[0]
	voiceName := req.VoiceActor.BaseVoice
	if voiceName == "" {
		voiceName = "Aoede"
	}
	stylePrompt := req.VoiceActor.StylePrompt
	if stylePrompt == "" {
		stylePrompt = "# AUDIO PROFILE: Default\n## THE SCENE: A professional studio"
	}

	ttsPrompt := fmt.Sprintf(`%s

### VARIATION SPECIFIC DIRECTION
Persona: %s
Subtext: %s
Technical: %s

#### TRANSCRIPT
%s`, stylePrompt, v.Persona, v.Subtext, v.TechnicalEnergy, v.Text)

	var ttsResp *genai.GenerateContentResponse
	maxRetries := MaxTTSRetries
	for attempt := 1; attempt <= maxRetries; attempt++ {
		currentModel := ttsModel
		if attempt == maxRetries {
			currentModel = "gemini-2.5-flash-tts"
			slog.Info("Falling back to backup TTS model", "variation", 0, "model", currentModel)
		}

		slog.Info("Generating TTS (One-Up)", 
			"voicePreset", req.VoiceActor.ShortName, 
			"model", currentModel, 
			"text", v.Text)
		ttsResp, err = client.Models.GenerateContent(ctx, currentModel,
			genai.Text(ttsPrompt), &genai.GenerateContentConfig{
				ResponseModalities: []string{"AUDIO"},
				SpeechConfig: &genai.SpeechConfig{
					VoiceConfig: &genai.VoiceConfig{
						PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
							VoiceName: voiceName,
						},
					},
				},
				SafetySettings: []*genai.SafetySetting{
					{
						Category:  genai.HarmCategoryHarassment,
						Threshold: genai.HarmBlockThresholdBlockNone,
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

		slog.Warn("TTS generation failed or returned empty", "attempt", attempt, "error", err, "blockReason", blockReason)
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	if err != nil || len(ttsResp.Candidates) == 0 || len(ttsResp.Candidates[0].Content.Parts) == 0 {
		slog.Error("All TTS generation attempts failed", "error", err)
		http.Error(w, "Failed to generate TTS audio", http.StatusInternalServerError)
		return
	}

	if len(ttsResp.Candidates) > 0 && len(ttsResp.Candidates[0].Content.Parts) > 0 {
		for _, part := range ttsResp.Candidates[0].Content.Parts {
			if part.InlineData != nil {
				audioData := part.InlineData.Data
				mimeType := part.InlineData.MIMEType

				if strings.HasPrefix(strings.ToLower(mimeType), "audio/l16") {
					audioData = addWavHeader(audioData, PCMDefaultSampleRate, PCMDefaultChannels, PCMDefaultBitDepth)
					mimeType = "audio/wav"
				}

				filename := generateFilename(mimeType)
				url, uploadErr := uploadToGCS(ctx, bucketName, filename, mimeType, audioData)
				if uploadErr != nil {
					slog.Error("Failed to upload audio to GCS", "error", uploadErr)
					http.Error(w, "Failed to upload audio to storage: "+uploadErr.Error(), http.StatusInternalServerError)
					return
				}
				v.Audio = url
				v.MimeType = mimeType
				break
			}
		}
	}

	variations[0] = v

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(variations[0])
}
