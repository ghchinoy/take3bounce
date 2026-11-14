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
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"go.opentelemetry.io/otel"
)

// addWavHeader dynamically constructs a valid 44-byte RIFF/WAVE header and prepends
// it to raw PCM byte arrays returned by APIs like Gemini TTS. This is required
// because browsers cannot play raw PCM streams natively via HTML5 audio tags.
func addWavHeader(rawData []byte, sampleRate uint32, numChannels uint16, bitsPerSample uint16) []byte {
	const headerLen = 44
	dataLen := uint32(len(rawData))
	fileSize := dataLen + 36
	byteRate := sampleRate * uint32(numChannels) * uint32(bitsPerSample/8)
	blockAlign := numChannels * (bitsPerSample / 8)

	out := make([]byte, headerLen+len(rawData))

	// RIFF/WAVE header
	copy(out[0:4], "RIFF")
	binary.LittleEndian.PutUint32(out[4:8], fileSize)
	copy(out[8:12], "WAVE")

	// fmt subchunk
	copy(out[12:16], "fmt ")
	binary.LittleEndian.PutUint32(out[16:20], 16) // Subchunk1Size (16 for PCM)
	binary.LittleEndian.PutUint16(out[20:22], 1)  // AudioFormat (1 for PCM)
	binary.LittleEndian.PutUint16(out[22:24], numChannels)
	binary.LittleEndian.PutUint32(out[24:28], sampleRate)
	binary.LittleEndian.PutUint32(out[28:32], byteRate)
	binary.LittleEndian.PutUint16(out[32:34], blockAlign)
	binary.LittleEndian.PutUint16(out[34:36], bitsPerSample)

	// data subchunk
	copy(out[36:40], "data")
	binary.LittleEndian.PutUint32(out[40:44], dataLen)
	copy(out[headerLen:], rawData)

	return out
}

// generateFilename creates a unique object key for Google Cloud Storage using
// the current unix timestamp and a random byte slice to prevent collisions.
// It infers the file extension from the provided MIME type.
func generateFilename(mimeType string) string {
	ext := ".mp3"
	switch {
	case strings.Contains(mimeType, "wav"):
		ext = ".wav"
	case strings.Contains(mimeType, "ogg"):
		ext = ".ogg"
	}

	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("take-%d-%x%s", time.Now().UnixNano(), b, ext)
}

// uploadToGCS uploads byte data to a Google Cloud Storage bucket. Crucially,
// it generates a pseudo-UUID and injects it into the object's metadata as a
// 'firebaseStorageDownloadTokens' key. This allows the file to be served
// publicly via the firebasestorage.googleapis.com endpoint without IAM auth,
// which is necessary for HTML5 <audio> tags.
func uploadToGCS(ctx context.Context, bucketName, filename, mimeType string, data []byte) (string, error) {
	ctx, span := otel.Tracer("threeup-orchestrator").Start(ctx, "GCS_Audio_Upload")
	defer span.End()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	// Generate a pseudo-UUID for the Firebase download token
	tb := make([]byte, FirebaseDownloadTokenLength)
	_, _ = rand.Read(tb)
	token := fmt.Sprintf("%x", tb)

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(filename)
	writer := obj.NewWriter(ctx)
	writer.ContentType = mimeType
	writer.Metadata = map[string]string{
		"firebaseStorageDownloadTokens": token,
	}

	if _, err := io.Copy(writer, bytes.NewReader(data)); err != nil {
		return "", fmt.Errorf("failed to write data to bucket: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close bucket writer: %w", err)
	}

	return fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media&token=%s", bucketName, filename, token), nil
}
