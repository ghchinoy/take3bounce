package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	recaptcha "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	recaptchapb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
)

// VerifyRecaptcha checks the token if RECAPTCHA_SITE_KEY is set.
// If it's not set, it returns true, allowing the request (Simple Mode).
func VerifyRecaptcha(ctx context.Context, token string, action string) (bool, error) {
	siteKey := os.Getenv("RECAPTCHA_SITE_KEY")
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

	if siteKey == "" {
		slog.Info("RECAPTCHA_SITE_KEY not set: Skipping validation (Simple Mode)")
		return true, nil
	}

	if token == "" {
		return false, fmt.Errorf("missing recaptcha token")
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
		return false, fmt.Errorf("unexpected action: %v", response.TokenProperties.Action)
	}

	if response.RiskAnalysis.Score < 0.5 {
		slog.Warn("ReCaptcha score too low", "score", response.RiskAnalysis.Score)
		return false, fmt.Errorf("score too low")
	}

	slog.Info("ReCaptcha validation passed", "score", response.RiskAnalysis.Score)
	return true, nil
}
