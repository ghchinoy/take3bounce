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

import "testing"

// TestNormalizeTags asserts observable behavior only, so it survives the
// planned Phase 3 refactor that hoists the regexes to package level.
func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		// Each alias -> canonical mapping.
		{"sigh", "hello [sigh] world", "hello [sighs] world"},
		{"laughing", "[laughing] ok", "[laughs] ok"},
		{"sarcasm", "sure [sarcasm]", "sure [sarcastic]"},
		{"whispering", "[whispering] quiet", "[whispers] quiet"},
		{"mischievous", "[mischievous] grin", "[mischievously] grin"},
		{"amazement", "[amazement] wow", "[amazed] wow"},
		{"excitement", "[excitement] yay", "[excited] yay"},

		// Case-insensitivity.
		{"uppercase sigh", "[SIGH]", "[sighs]"},
		{"mixed case laughing", "[LaUgHiNg]", "[laughs]"},

		// Multiple tags in one string.
		{"multiple tags", "[sigh] and [laughing]", "[sighs] and [laughs]"},

		// Already-canonical tags are left untouched (no over-matching).
		{"canonical sighs untouched", "[sighs]", "[sighs]"},
		{"canonical laughs untouched", "[laughs]", "[laughs]"},

		// No over-matching: substrings of words are not tags.
		{"word sigh not a tag", "she gave a sigh", "she gave a sigh"},
		{"no brackets", "laughing out loud", "laughing out loud"},

		// Untouched, unrelated tags pass through.
		{"unknown tag untouched", "[pause] here", "[pause] here"},

		// Empty input.
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeTags(tt.in); got != tt.want {
				t.Errorf("normalizeTags(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
