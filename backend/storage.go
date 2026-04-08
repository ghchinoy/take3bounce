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
)

// addWavHeader dynamically constructs a valid RIFF/WAVE header and prepends
// it to raw PCM byte arrays returned by APIs like Gemini TTS. This is required
// because browsers cannot play raw PCM streams natively via HTML5 audio tags.
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

// generateFilename creates a unique object key for Google Cloud Storage using
// the current unix timestamp and a random byte slice to prevent collisions.
// It infers the file extension from the provided MIME type.
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

// uploadToGCS uploads byte data to a Google Cloud Storage bucket. Crucially,
// it generates a pseudo-UUID and injects it into the object's metadata as a
// 'firebaseStorageDownloadTokens' key. This allows the file to be served
// publicly via the firebasestorage.googleapis.com endpoint without IAM auth,
// which is necessary for HTML5 <audio> tags.
func uploadToGCS(ctx context.Context, bucketName, filename, mimeType string, data []byte) (string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create storage client: %v", err)
	}
	defer client.Close()

	// Generate a pseudo-UUID for the Firebase download token
	tb := make([]byte, FirebaseDownloadTokenLength)
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
