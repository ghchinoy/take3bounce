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

const (
	// MaxTTSRetries defines the maximum number of attempts to make when calling
	// the Gemini TTS API, specifically to handle PROHIBITED_CONTENT blocks.
	MaxTTSRetries = 3

	// Audio Specifications for Gemini 3.1 Flash TTS
	// The API returns raw audio/l16 (PCM) which requires a RIFF/WAVE header
	// to be dynamically constructed for browser playback.
	PCMDefaultSampleRate uint32 = 24000
	PCMDefaultChannels   uint16 = 1
	PCMDefaultBitDepth   uint16 = 16

	// FirebaseDownloadTokenLength defines the byte length for the pseudo-UUID
	// used to bypass IAM authentication for direct browser audio playback.
	FirebaseDownloadTokenLength = 16
)
