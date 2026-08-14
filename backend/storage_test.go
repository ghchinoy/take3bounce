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
	"encoding/binary"
	"strings"
	"testing"
)

func TestAddWavHeader(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		sampleRate    uint32
		numChannels   uint16
		bitsPerSample uint16
	}{
		{
			name:          "typical mono 24kHz 16-bit PCM",
			data:          []byte{0x01, 0x02, 0x03, 0x04},
			sampleRate:    24000,
			numChannels:   1,
			bitsPerSample: 16,
		},
		{
			name:          "stereo 48kHz 16-bit PCM",
			data:          []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
			sampleRate:    48000,
			numChannels:   2,
			bitsPerSample: 16,
		},
		{
			name:          "empty PCM payload",
			data:          []byte{},
			sampleRate:    16000,
			numChannels:   1,
			bitsPerSample: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := addWavHeader(tt.data, tt.sampleRate, tt.numChannels, tt.bitsPerSample)

			// Header is always 44 bytes, followed by the raw data.
			const headerLen = 44
			if len(out) != headerLen+len(tt.data) {
				t.Fatalf("total length = %d, want %d", len(out), headerLen+len(tt.data))
			}

			// The raw PCM data must be preserved verbatim after the header.
			if got := out[headerLen:]; string(got) != string(tt.data) {
				t.Errorf("payload not preserved: got %v, want %v", got, tt.data)
			}

			// ASCII markers at their canonical offsets.
			assertMarker(t, out, 0, "RIFF")
			assertMarker(t, out, 8, "WAVE")
			assertMarker(t, out, 12, "fmt ")
			assertMarker(t, out, 36, "data")

			dataLen := uint32(len(tt.data))

			// fileSize field (bytes 4-8) == len(data) + 36.
			if got := binary.LittleEndian.Uint32(out[4:8]); got != dataLen+36 {
				t.Errorf("fileSize = %d, want %d", got, dataLen+36)
			}

			// Subchunk1Size (bytes 16-20) == 16 for PCM.
			if got := binary.LittleEndian.Uint32(out[16:20]); got != 16 {
				t.Errorf("Subchunk1Size = %d, want 16", got)
			}

			// AudioFormat (bytes 20-22) == 1 for PCM.
			if got := binary.LittleEndian.Uint16(out[20:22]); got != 1 {
				t.Errorf("AudioFormat = %d, want 1", got)
			}

			// NumChannels (bytes 22-24).
			if got := binary.LittleEndian.Uint16(out[22:24]); got != tt.numChannels {
				t.Errorf("NumChannels = %d, want %d", got, tt.numChannels)
			}

			// SampleRate (bytes 24-28).
			if got := binary.LittleEndian.Uint32(out[24:28]); got != tt.sampleRate {
				t.Errorf("SampleRate = %d, want %d", got, tt.sampleRate)
			}

			// ByteRate (bytes 28-32) == sampleRate * numChannels * bitsPerSample/8.
			wantByteRate := tt.sampleRate * uint32(tt.numChannels) * uint32(tt.bitsPerSample/8)
			if got := binary.LittleEndian.Uint32(out[28:32]); got != wantByteRate {
				t.Errorf("ByteRate = %d, want %d", got, wantByteRate)
			}

			// BlockAlign (bytes 32-34) == numChannels * bitsPerSample/8.
			wantBlockAlign := tt.numChannels * (tt.bitsPerSample / 8)
			if got := binary.LittleEndian.Uint16(out[32:34]); got != wantBlockAlign {
				t.Errorf("BlockAlign = %d, want %d", got, wantBlockAlign)
			}

			// BitsPerSample (bytes 34-36).
			if got := binary.LittleEndian.Uint16(out[34:36]); got != tt.bitsPerSample {
				t.Errorf("BitsPerSample = %d, want %d", got, tt.bitsPerSample)
			}

			// data subchunk size (bytes 40-44) == len(data).
			if got := binary.LittleEndian.Uint32(out[40:44]); got != dataLen {
				t.Errorf("data subchunk size = %d, want %d", got, dataLen)
			}
		})
	}
}

func assertMarker(t *testing.T, out []byte, offset int, want string) {
	t.Helper()
	if got := string(out[offset : offset+len(want)]); got != want {
		t.Errorf("marker at offset %d = %q, want %q", offset, got, want)
	}
}

func TestGenerateFilename(t *testing.T) {
	extTests := []struct {
		name     string
		mimeType string
		wantExt  string
	}{
		{"wav mime", "audio/wav", ".wav"},
		{"wav substring", "audio/x-wav", ".wav"},
		{"ogg mime", "audio/ogg", ".ogg"},
		{"mp3 mime defaults", "audio/mpeg", ".mp3"},
		{"unknown defaults to mp3", "application/octet-stream", ".mp3"},
		{"empty defaults to mp3", "", ".mp3"},
	}

	for _, tt := range extTests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateFilename(tt.mimeType)
			if !strings.HasPrefix(got, "take-") {
				t.Errorf("filename %q does not have %q prefix", got, "take-")
			}
			if !strings.HasSuffix(got, tt.wantExt) {
				t.Errorf("filename %q does not have extension %q", got, tt.wantExt)
			}
		})
	}

	t.Run("unique across calls", func(t *testing.T) {
		const n = 1000
		seen := make(map[string]struct{}, n)
		for i := 0; i < n; i++ {
			name := generateFilename("audio/wav")
			if _, dup := seen[name]; dup {
				t.Fatalf("duplicate filename generated: %q", name)
			}
			seen[name] = struct{}{}
		}
	})
}
