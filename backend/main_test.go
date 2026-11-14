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
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestParseOriginSet covers the comma-separated ALLOWED_ORIGINS parser:
// trimming, empty-entry dropping, and the unset/blank -> empty-set default.
func TestParseOriginSet(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string // origins expected present; length is also the set size
	}{
		{"empty_string", "", nil},
		{"whitespace_only", "   ", nil},
		{"commas_only", ",, ,", nil},
		{"single_origin", "https://app.run.app", []string{"https://app.run.app"}},
		{
			"two_origins",
			"https://app.run.app,http://localhost:5173",
			[]string{"https://app.run.app", "http://localhost:5173"},
		},
		{
			"whitespace_around_entries",
			" https://app.run.app , http://localhost:5173 ",
			[]string{"https://app.run.app", "http://localhost:5173"},
		},
		{
			"trailing_comma",
			"https://app.run.app,",
			[]string{"https://app.run.app"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOriginSet(tt.raw)
			if len(got) != len(tt.want) {
				t.Fatalf("parseOriginSet(%q) size = %d, want %d (%v)", tt.raw, len(got), len(tt.want), got)
			}
			for _, o := range tt.want {
				if _, ok := got[o]; !ok {
					t.Errorf("parseOriginSet(%q) missing origin %q", tt.raw, o)
				}
			}
		})
	}
}

// TestCORSAllowedOriginEchoed verifies that a listed origin is echoed back
// exactly, with Vary: Origin and the trimmed method list, and the request proceeds.
func TestCORSAllowedOriginEchoed(t *testing.T) {
	var reached bool
	h := testMiddleware()(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/api/variations", nil)
	req.Header.Set("Origin", allowedOrigin)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("newCORSMiddleware(Origin=%q) Access-Control-Allow-Origin = %q, want %q", allowedOrigin, got, allowedOrigin)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("newCORSMiddleware(Origin=%q) Vary = %q, want %q", allowedOrigin, got, "Origin")
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "POST, GET, OPTIONS" {
		t.Errorf("newCORSMiddleware(Origin=%q) Access-Control-Allow-Methods = %q, want %q", allowedOrigin, got, "POST, GET, OPTIONS")
	}
	if !reached {
		t.Errorf("newCORSMiddleware(Origin=%q) did not reach next handler", allowedOrigin)
	}
}

// TestCORSDisallowedOriginNoHeader verifies that an unlisted origin gets no
// ACAO header, while the request itself is not blocked by the server.
func TestCORSDisallowedOriginNoHeader(t *testing.T) {
	const disallowed = "https://evil.example.com"
	var reached bool
	h := testMiddleware()(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/api/variations", nil)
	req.Header.Set("Origin", disallowed)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("newCORSMiddleware(Origin=%q) Access-Control-Allow-Origin = %q, want empty", disallowed, got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "" {
		t.Errorf("newCORSMiddleware(Origin=%q) Access-Control-Allow-Methods = %q, want empty", disallowed, got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("newCORSMiddleware(Origin=%q) Vary = %q, want %q", disallowed, got, "Origin")
	}
	if !reached {
		t.Errorf("newCORSMiddleware(Origin=%q) did not run next handler", disallowed)
	}
}

// TestCORSVaryAlwaysEmitted verifies that Vary: Origin is present on every
// response the middleware handles, including a same-origin request with no Origin header.
func TestCORSVaryAlwaysEmitted(t *testing.T) {
	var reached bool
	h := testMiddleware()(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/api/variations", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("newCORSMiddleware(Origin=\"\") Vary = %q, want %q", got, "Origin")
	}
	if !reached {
		t.Error("newCORSMiddleware(Origin=\"\") blocked same-origin request, want pass-through")
	}
}

// TestCORSNeverWildcard verifies that no code path ever emits "*" as the
// allowed origin, including when the incoming Origin literally is "*".
func TestCORSNeverWildcard(t *testing.T) {
	h := testMiddleware()(okHandler(new(bool)))

	for _, origin := range []string{allowedOrigin, "https://evil.example.com", "*"} {
		req := httptest.NewRequest(http.MethodPost, "/api/variations", nil)
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got == "*" {
			t.Errorf("newCORSMiddleware(Origin=%q) Access-Control-Allow-Origin = %q, want non-wildcard", origin, got)
		}
	}
}

// TestCORSPreflightAllowed verifies that OPTIONS from an allowed origin returns
// 204 with the CORS headers and does not fall through to the next handler.
func TestCORSPreflightAllowed(t *testing.T) {
	var reached bool
	h := testMiddleware()(okHandler(&reached))

	req := httptest.NewRequest(http.MethodOptions, "/api/variations", nil)
	req.Header.Set("Origin", allowedOrigin)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("newCORSMiddleware(OPTIONS, Origin=%q) status = %d, want %d", allowedOrigin, rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("newCORSMiddleware(OPTIONS, Origin=%q) Access-Control-Allow-Origin = %q, want %q", allowedOrigin, got, allowedOrigin)
	}
	if reached {
		t.Errorf("newCORSMiddleware(OPTIONS, Origin=%q) reached next handler, want short-circuit", allowedOrigin)
	}
}

// TestCORSPreflightDisallowed verifies that OPTIONS from a disallowed origin
// returns 204 with no CORS headers.
func TestCORSPreflightDisallowed(t *testing.T) {
	const disallowed = "https://evil.example.com"
	h := testMiddleware()(okHandler(new(bool)))

	req := httptest.NewRequest(http.MethodOptions, "/api/variations", nil)
	req.Header.Set("Origin", disallowed)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("newCORSMiddleware(OPTIONS, Origin=%q) status = %d, want %d", disallowed, rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("newCORSMiddleware(OPTIONS, Origin=%q) Access-Control-Allow-Origin = %q, want empty", disallowed, got)
	}
}

// TestCORSSameOriginPassthrough verifies that a request with no Origin header
// is never blocked and gets no CORS headers.
func TestCORSSameOriginPassthrough(t *testing.T) {
	var reached bool
	h := newCORSMiddleware(parseOriginSet(""))(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/api/variations", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !reached {
		t.Error("newCORSMiddleware(allowlist=\"\", Origin=\"\") blocked same-origin request, want pass-through")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("newCORSMiddleware(allowlist=\"\", Origin=\"\") Access-Control-Allow-Origin = %q, want empty", got)
	}
}

const allowedOrigin = "https://app.run.app"

func testMiddleware() func(http.Handler) http.Handler {
	return newCORSMiddleware(parseOriginSet(allowedOrigin + ",http://localhost:5173"))
}

// okHandler is the "next" handler; it records that it was reached.
func okHandler(reached *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		*reached = true
		w.WriteHeader(http.StatusOK)
	})
}
