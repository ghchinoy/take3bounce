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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	recaptcha "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	recaptchapb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
)

// VerifyRecaptcha checks the token if RECAPTCHA_SITE_KEY is set.
// If it is not set, it returns true, allowing the request (Simple Mode).
func VerifyRecaptcha(ctx context.Context, token, action string) (bool, error) {
	siteKey := os.Getenv("RECAPTCHA_SITE_KEY")
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

	if siteKey == "" {
		slog.Info("RECAPTCHA_SITE_KEY not set: Skipping validation (Simple Mode)")
		return true, nil
	}

	if token == "" {
		return false, errors.New("missing recaptcha token")
	}

	client, err := recaptcha.NewClient(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to create recaptcha client: %w", err)
	}
	defer client.Close()

	event := &recaptchapb.Event{
		Token:   token,
		SiteKey: siteKey,
	}

	request := &recaptchapb.CreateAssessmentRequest{
		Assessment: &recaptchapb.Assessment{
			Event: event,
		},
		Parent: fmt.Sprintf("projects/%s", projectID),
	}

	response, err := client.CreateAssessment(ctx, request)
	if err != nil {
		return false, fmt.Errorf("failed to create assessment: %w", err)
	}

	if !response.TokenProperties.Valid {
		slog.Warn("ReCaptcha invalid token", "reason", response.TokenProperties.InvalidReason)
		return false, fmt.Errorf("invalid token: %v", response.TokenProperties.InvalidReason)
	}

	if response.TokenProperties.Action != action {
		slog.Warn("ReCaptcha action mismatch", "expected", action, "got", response.TokenProperties.Action)
		return false, fmt.Errorf("unexpected action %q, want %q", response.TokenProperties.Action, action)
	}

	if response.RiskAnalysis.Score < 0.5 {
		slog.Warn("ReCaptcha score too low", "score", response.RiskAnalysis.Score)
		return false, errors.New("score too low")
	}

	slog.Info("ReCaptcha validation passed", "score", response.RiskAnalysis.Score)
	return true, nil
}
