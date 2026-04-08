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
