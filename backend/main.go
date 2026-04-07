package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

type GenerateRequest struct {
	Text string `json:"text"`
}

type Variation struct {
	Take            string `json:"take"`
	Persona         string `json:"persona"`
	Subtext         string `json:"subtext"`
	TechnicalEnergy string `json:"technicalEnergy"`
	Text            string `json:"text"`
	Audio           string `json:"audio,omitempty"`    // URL to GCS audio file
	MimeType        string `json:"mimeType,omitempty"` // audio mime type
}

func addWavHeader(rawData []byte, sampleRate uint32, numChannels uint16, bitsPerSample uint16) []byte {
	// Calculate sizes
	dataLen := uint32(len(rawData))
	fileSize := dataLen + 36

	header := new(bytes.Buffer)

	// RIFF/WAVE header
	header.WriteString("RIFF")
	binary.Write(header, binary.LittleEndian, fileSize)
	header.WriteString("WAVE")

	// fmt subchunk
	header.WriteString("fmt ")
	binary.Write(header, binary.LittleEndian, uint32(16)) // Subchunk1Size (16 for PCM)
	binary.Write(header, binary.LittleEndian, uint16(1))  // AudioFormat (1 for PCM)
	binary.Write(header, binary.LittleEndian, numChannels)
	binary.Write(header, binary.LittleEndian, sampleRate)
	byteRate := sampleRate * uint32(numChannels) * uint32(bitsPerSample/8)
	binary.Write(header, binary.LittleEndian, byteRate)
	blockAlign := numChannels * (bitsPerSample / 8)
	binary.Write(header, binary.LittleEndian, blockAlign)
	binary.Write(header, binary.LittleEndian, bitsPerSample)

	// data subchunk
	header.WriteString("data")
	binary.Write(header, binary.LittleEndian, dataLen)

	return append(header.Bytes(), rawData...)
}

func generateFilename(mimeType string) string {
	ext := ".mp3"
	if strings.Contains(mimeType, "wav") {
		ext = ".wav"
	} else if strings.Contains(mimeType, "ogg") {
		ext = ".ogg"
	}

	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("take-%d-%x%s", time.Now().UnixNano(), b, ext)
}

func uploadToGCS(ctx context.Context, bucketName, filename, mimeType string, data []byte) (string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create storage client: %v", err)
	}
	defer client.Close()

	// Generate a pseudo-UUID for the Firebase download token
	tb := make([]byte, 16)
	rand.Read(tb)
	token := fmt.Sprintf("%x", tb)

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(filename)
	writer := obj.NewWriter(ctx)
	writer.ContentType = mimeType
	writer.Metadata = map[string]string{
		"firebaseStorageDownloadTokens": token,
	}

	if _, err := io.Copy(writer, bytes.NewReader(data)); err != nil {
		return "", fmt.Errorf("failed to write data to bucket: %v", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close bucket writer: %v", err)
	}

	return fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media&token=%s", bucketName, filename, token), nil
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	_ = godotenv.Load()

	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if project == "" || location == "" {
		slog.Error("FATAL: GOOGLE_CLOUD_PROJECT and GOOGLE_CLOUD_LOCATION must be set in the environment")
		os.Exit(1)
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/variations", handleVariations).Methods("POST", "OPTIONS")

	// Serve static files from the frontend build
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./dist")))

	r.Use(enableCORS)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Backend server listening", "port", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

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

	prompt := fmt.Sprintf(`You are a VO director. Create a "Three-Up" based on the following text.
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
%s`, req.Text)

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

			ttsResp, err := client.Models.GenerateContent(ctx, ttsModel,
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

			if err != nil {
				slog.Error("TTS error for variation", "index", idx, "error", err)
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
