#!/usr/bin/env bash
# Run ON THE MAC HOST, from the take3bounce repo root:  bash docs/commands/01-gemini-3-8-tts-migration-epic.sh
# Container agents can't write bd (BEADS_DOLT_AUTO_START=false); this drops the
# bd writes for the host, which owns the single Dolt lock.
#
# Context: Gemini 3.8 TTS migration (2026-10-07). take3bounce goes 3.8-ONLY.
#   Shared analysis: scream/docs/GEMINI_3_8_TTS_MIGRATION.md (sections 4 + 5.2)
#   Evidence:        scream/prototypes/tts-38-probe/
# NOT idempotent for creates (IDs are allocated by bd); run once.

set -euo pipefail
test -d .beads || { echo "Run from the take3bounce repo root on the Mac."; exit 1; }
ACTOR="${BEADS_ACTOR:-agent:container:implementer}"
SPEC="../scream/docs/GEMINI_3_8_TTS_MIGRATION.md"

mk() { # mk <type> <prio> <title> <desc> [parent]
  local args=(create "$3" -t "$1" -p "$2" -d "$4" --spec-id "$SPEC" --actor "$ACTOR" --silent)
  [ -n "${5:-}" ] && args+=(--parent "$5")
  bd "${args[@]}"
}

EPIC=$(mk epic 1 "Epic: Migrate to Gemini 3.8 TTS (3.8-only)" \
"Move TTS from gemini-3.1-flash-tts-preview to gemini-3.8-flash-tts (default) / gemini-3.8-flash-lite-tts. No 3.1 code path. See $SPEC sections 4 and 5.2. Done when: Three-Up, One-Up, Sandbox and Showcase all produce 3.8 audio with per-take speechMetadata.style and the 3.8 tag vocabulary; docs/copy updated; deployed.")
echo "epic $EPIC"

SDK=$(mk task 1 "Upgrade google.golang.org/genai v1.52.1 -> v1.72.0" \
"backend/go.mod. v1.72.0 exposes Part.SpeechMetadata and VoiceConfig.Voice natively (verified in scream/prototypes/tts-38-probe/gosdk). Also clean go.mod '// indirect' markers via go mod tidy. Acceptance: go build/vet/test green." "$EPIC")

CLIENT=$(mk task 1 "Single synthesizeTake() helper: global TTS client, parts+style, voice, WAV-aware" \
"Collapse the three copy-pasted TTS blocks (generator.go:248-264, 416-432, 632-648) into one helper. New env GEMINI_TTS_LOCATION (default global); reuse a client instead of NewClient per request. Request: one part {Text, SpeechMetadata{Style}}, SpeechConfig.VoiceConfig.Voice. Unary returns audio/wav: keep the audio/l16 check so addWavHeader is skipped (verify). Fallback model = the other 3.8 model (TTS_FALLBACK_MODEL). Safety: all BLOCK_NONE accepted on 3.8 (probe F1/F2); keep Harassment override or drop. Fix linear-vs-exponential backoff comment. Acceptance: unit test with fake client asserting request shape." "$EPIC")

DIRECTOR=$(mk task 1 "Director prompt v2: per-take short style + 3.8 tags in text" \
"generator.go:33-80 (OneUp + Director templates), tags.go. JSON per take gains 'style' (short, <=~15 words, e.g. 'wry, unhurried, half-smiling'); 'text' contains only the 3.8 vocabulary (<sigh> <laugh> <chuckle> <gasp> <short pause> <long pause> ...; see doc 4.2), CAPS for emphasis, disfluencies as words. Wildcard drama via style (probe B5: [whispering] inline is spoken aloud). Remove Full/Enhanced A/B (single vocabulary) and the telemetry attribute or repurpose it. Note style can lose to text sentiment (L2c): prompt for specific, physical style words. Acceptance: parser tests for the new schema; manual listen of 3 scripts." "$EPIC")

TTSPROMPT=$(mk task 1 "Drop Markdown TTS wrapper; style = stack(actor baseline, take style)" \
"generator.go:212-220 (x3). Text = transcript only. Style = dedup-join of actor.baselineStyle and take.style (pattern: scream stackDirectives), soft cap ~120 chars. Acceptance: no '# AUDIO PROFILE' / 'TRANSCRIPT' strings reach the TTS request." "$EPIC")

NORM=$(mk task 2 "Tag normaliser: legacy [square] -> 3.8 <angle>; emotion/style tags -> style or stripped" \
"backend normalizeTags (generator.go:720-744) + frontend normalizeTextTags (audio-tags.ts:265-277). Map per doc 4.2 ([sighs]-><sigh>, [laughing]-><laugh>, [uhm]->uhm, [medium pause]-><short pause>, [whispering]/[sarcastic]/emotions -> removed from text). Update TestNormalizeTags. Acceptance: user-edited Regenerate text with legacy tags produces no spoken tag words." "$EPIC")

PERS=$(mk task 2 "Voice actors for 3.8: voice + short baseline style (dedupe preset list)" \
"app-main.ts:53-74 and oneup-app.ts:28-33 (dedupe into presets.ts). Shape {voice, baselineStyle}. Accents can't live in style: Auntie Mae -> en-au-* library voice (e.g. en-au-training-6, 57y cooking show host); Glitch (Estuary) and Jaz (Brixton) -> no London accent in library: Voice design or Manchester/Winchester library voice; Cosmic Curator -> Orus + 'reverent, hushed awe'. Designed voices are per-project and expire 1y after last use, so 'Run on Google Cloud' deploys need a library fallback or an ensure-voices bootstrap. Acceptance: each preset audibly matches its card." "$EPIC")

UI=$(mk task 2 "UI: Showcase + tag editor pills + Sandbox style field on 3.8 vocabulary" \
"showcase-app.ts (one sample per tag -> 3.8 tags), audio-tags.ts defaultTags/extendedTags/allTags, variation-card.ts pills, sandbox-app.ts (add style input), display per-take style on the card." "$EPIC")

DOCS=$(mk task 3 "Docs, config and copy: 3.8 model names, pricing, safety note" \
"README (models, cost table), GEMINI.md:39-49, docs/ARCHITECTURE.md:15-21 (PCM + Harassment-only notes are 3.1-era), app.json, backend/.env.sample, scripts/deploy.sh (also forward TTS_FALLBACK_MODEL and GEMINI_TTS_LOCATION), UI copy app-main.ts:442, oneup-app.ts:380, deploy-modal.ts:83-86, tools/test_tts_safety*.go." "$EPIC")

RECAP=$(mk bug 2 "Retry/Regenerate omit recaptchaToken (break when RECAPTCHA_SITE_KEY set)" \
"variation-card.ts:164-168 and 197-201 POST /api/retry-audio without recaptchaToken; backend expects action generate_retry. Found during the 3.8 review.")

for c in "$CLIENT" "$DIRECTOR"; do bd dep add "$c" "$SDK" --actor "$ACTOR" || true; done
bd dep add "$TTSPROMPT" "$CLIENT" --actor "$ACTOR" || true
bd dep add "$TTSPROMPT" "$DIRECTOR" --actor "$ACTOR" || true
bd dep add "$PERS" "$TTSPROMPT" --actor "$ACTOR" || true
bd dep add "$UI" "$NORM" --actor "$ACTOR" || true
bd dep add "$DOCS" "$TTSPROMPT" --actor "$ACTOR" || true

bd export -o .beads/issues.jsonl
git status --short .beads/issues.jsonl
echo "Suggested: git add .beads/issues.jsonl && git commit -m 'bd: Gemini 3.8 TTS migration epic'"
