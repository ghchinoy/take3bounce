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

// VoiceActor represents a virtual persona for the audio generation.
// It dictates the base voice model to use and the underlying style instructions
// (Director's Notes) that influence the performance.
type VoiceActor struct {
	ShortName   string `json:"shortName"`
	BaseVoice   string `json:"baseVoice"`
	StylePrompt string `json:"stylePrompt"`
}

// GenerateRequest represents the incoming JSON payload from the frontend
// when a user clicks 'Take three, on the bounce' to orchestrate multiple takes.
type GenerateRequest struct {
	Text           string     `json:"text"`
	VoiceActor     VoiceActor `json:"voiceActor"`
	RecaptchaToken string     `json:"recaptchaToken"`
}

// RetryAudioRequest represents the incoming JSON payload from the frontend
// for regenerating a single audio take, either due to a safety filter failure
// or manual user regeneration.
type RetryAudioRequest struct {
	Variation      Variation  `json:"variation"`
	VoiceActor     VoiceActor `json:"voiceActor"`
	RecaptchaToken string     `json:"recaptchaToken"`
}

// Variation represents a single generated 'Take' (e.g., Safe, Pushed, Wildcard).
// It contains the LLM-generated direction (Persona, Subtext, TechnicalEnergy)
// as well as the script text and the resulting synthesized audio URL.
type Variation struct {
	Take            string `json:"take"`
	Persona         string `json:"persona"`
	Subtext         string `json:"subtext"`
	TechnicalEnergy string `json:"technicalEnergy"`
	Text            string `json:"text"`
	Audio           string `json:"audio,omitempty"`    // URL to GCS audio file
	MimeType        string `json:"mimeType,omitempty"` // audio mime type
}

// GenerateOneRequest represents the incoming JSON payload for the One-Up generator.
type GenerateOneRequest struct {
	Text           string     `json:"text"`
	VoiceActor     VoiceActor `json:"voiceActor"`
	ReadingTone    string     `json:"readingTone"`
	RecaptchaToken string     `json:"recaptchaToken"`
}
