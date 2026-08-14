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
		{"empty string", "", nil},
		{"whitespace only", "   ", nil},
		{"commas only", ",, ,", nil},
		{"single origin", "https://app.run.app", []string{"https://app.run.app"}},
		{
			"two origins",
			"https://app.run.app,http://localhost:5173",
			[]string{"https://app.run.app", "http://localhost:5173"},
		},
		{
			"whitespace around entries",
			" https://app.run.app , http://localhost:5173 ",
			[]string{"https://app.run.app", "http://localhost:5173"},
		},
		{
			"trailing comma",
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

const allowedOrigin = "https://app.run.app"

func testMiddleware() func(http.Handler) http.Handler {
	return newCORSMiddleware(parseOriginSet(allowedOrigin + ",http://localhost:5173"))
}

// okHandler is the "next" handler; it records that it was reached.
func okHandler(reached *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*reached = true
		w.WriteHeader(http.StatusOK)
	})
}

// TestCORSAllowedOriginEchoed: a listed origin is echoed back exactly, with
// Vary: Origin and the trimmed method list, and the request still proceeds.
func TestCORSAllowedOriginEchoed(t *testing.T) {
	var reached bool
	h := testMiddleware()(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/api/variations", nil)
	req.Header.Set("Origin", allowedOrigin)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, allowedOrigin)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want %q", got, "Origin")
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "POST, GET, OPTIONS" {
		t.Errorf("Access-Control-Allow-Methods = %q, want %q", got, "POST, GET, OPTIONS")
	}
	if !reached {
		t.Error("next handler was not reached for allowed non-preflight request")
	}
}

// TestCORSDisallowedOriginNoHeader: an unlisted origin gets no ACAO header, but
// the request is not blocked (CORS is enforced by the browser, not the server).
func TestCORSDisallowedOriginNoHeader(t *testing.T) {
	var reached bool
	h := testMiddleware()(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/api/variations", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty for disallowed origin", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "" {
		t.Errorf("Access-Control-Allow-Methods = %q, want empty for disallowed origin", got)
	}
	// Vary: Origin must be emitted even for a disallowed origin: the response
	// still varies by Origin (correct HTTP caching semantics).
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want %q for disallowed origin", got, "Origin")
	}
	if !reached {
		t.Error("next handler should still run for disallowed origin (server does not block)")
	}
}

// TestCORSVaryAlwaysEmitted: Vary: Origin is present on every response the
// middleware handles, including a same-origin request with no Origin header,
// so shared caches never serve a CORS response keyed without Origin.
func TestCORSVaryAlwaysEmitted(t *testing.T) {
	var reached bool
	h := testMiddleware()(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/api/variations", nil)
	// No Origin header set (same-origin request).
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want %q for absent Origin", got, "Origin")
	}
	if !reached {
		t.Error("same-origin request must not be blocked")
	}
}

// TestCORSNeverWildcard: no code path ever emits "*" as the allowed origin,
// including when the incoming Origin literally is "*".
func TestCORSNeverWildcard(t *testing.T) {
	h := testMiddleware()(okHandler(new(bool)))

	for _, origin := range []string{allowedOrigin, "https://evil.example.com", "*"} {
		req := httptest.NewRequest(http.MethodPost, "/api/variations", nil)
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got == "*" {
			t.Errorf("Origin %q produced wildcard Access-Control-Allow-Origin", origin)
		}
	}
}

// TestCORSPreflightAllowed: OPTIONS from an allowed origin returns 204 with the
// CORS headers and does NOT fall through to the next handler.
func TestCORSPreflightAllowed(t *testing.T) {
	var reached bool
	h := testMiddleware()(okHandler(&reached))

	req := httptest.NewRequest(http.MethodOptions, "/api/variations", nil)
	req.Header.Set("Origin", allowedOrigin)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("preflight Access-Control-Allow-Origin = %q, want %q", got, allowedOrigin)
	}
	if reached {
		t.Error("preflight OPTIONS should not reach the next handler")
	}
}

// TestCORSPreflightDisallowed: OPTIONS from a disallowed origin still returns
// 204 (a valid preflight response) but carries no CORS headers, so the browser
// rejects the cross-origin call.
func TestCORSPreflightDisallowed(t *testing.T) {
	h := testMiddleware()(okHandler(new(bool)))

	req := httptest.NewRequest(http.MethodOptions, "/api/variations", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("disallowed preflight Access-Control-Allow-Origin = %q, want empty", got)
	}
}

// TestCORSSameOriginPassthrough: a request with no Origin header (same-origin,
// the production default) is never blocked and gets no CORS headers.
func TestCORSSameOriginPassthrough(t *testing.T) {
	var reached bool
	// Empty allowlist == ALLOWED_ORIGINS unset (default production posture).
	h := newCORSMiddleware(parseOriginSet(""))(okHandler(&reached))

	req := httptest.NewRequest(http.MethodPost, "/api/variations", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !reached {
		t.Error("same-origin request (no Origin header) must not be blocked")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("same-origin request got Access-Control-Allow-Origin = %q, want empty", got)
	}
}
