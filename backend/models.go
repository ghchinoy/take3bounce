package main

type VoiceActor struct {
	ShortName   string `json:"shortName"`
	BaseVoice   string `json:"baseVoice"`
	StylePrompt string `json:"stylePrompt"`
}

type GenerateRequest struct {
	Text       string     `json:"text"`
	VoiceActor VoiceActor `json:"voiceActor"`
}

type RetryAudioRequest struct {
	Variation  Variation  `json:"variation"`
	VoiceActor VoiceActor `json:"voiceActor"`
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