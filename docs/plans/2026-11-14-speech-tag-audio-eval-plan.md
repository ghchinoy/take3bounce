# Speech Tag & Audio Direction Evaluation Plan (`mizan` + `mizan-templates`)

**Status:** Proposed (`P3`)
**Date:** 2026-11-15
**Related Repositories:** `https://github.com/ghchinoy/mizan`, `https://github.com/ghchinoy/mizan-templates`

---

## 1. Problem Statement

The Three-Up Audio Orchestrator relies on two layers of AI direction to synthesize distinct voiceover takes:
1. **Turn-level acting direction** (`persona`, `subtext`, and in Gemini 3.8 TTS, `speechMetadata.style`).
2. **Inline paralinguistic speech tags** (`[laughs]`, `[sighs]`, `[whispers]`, `[short pause]` in Gemini 3.1 TTS; `<laugh>`, `<sigh>`, `<whispers>`, `<pant>`, `<short pause>`, `<long pause>`, and `ALL-CAPS` emphasis in Gemini 3.8 TTS).

Currently, `take3bounce` has **no automated evaluations** that verify whether:
- Inline vocal tags produce actual acoustic vocalizations (e.g., an audible sigh or pause) at the marked position in the synthesized `.wav` output.
- Inline tags or stage directions are accidentally spoken aloud verbatim by the TTS engine (a critical regression mode when migrating from Gemini 3.1 TTS to Gemini 3.8 TTS, where square-bracket emotion tags like `[whispering]` or `[sarcastic]` are read aloud as literal words).
- Sustained style directives (`persona`, `subtext`, `speechMetadata.style`) audibly shape the actor's prosody, pacing, and emotional tone.
- The three generated takes (`Take 1: Safe`, `Take 2: Pushed`, `Take 3: Wildcard`) are acoustically distinct and follow the intended 1-2-3 energy ramp.

---

## 2. Existing Capabilities in `mizan` & `mizan-templates`

`mizan` (`v0.4.0+`) supports multimodal `audio` + `text` evaluation (`pointwise`, `rubric`, `custom_schema`, `pairwise`, `boul`, `choice`, `score`) using `gemini-3.5-flash` as an audio-native judge.

The existing `packs/audio/` domain pack in `mizan-templates` includes 4 general speech templates:
- `audio/speech-clarity-pointwise` (`pointwise`) — acoustic intelligibility and pronunciation.
- `audio/transcript-fidelity-pointwise` (`pointwise`) — word-for-word match against plain reference text.
- `audio/tts-naturalness-comparison` (`pairwise`) — A/B comparison of two TTS clips for natural prosody.
- `audio/voice-talent-rubric` (`rubric`) — general studio voiceover quality (1–5 scale).

**Gap:** None of the existing templates inspect **inline speech-tag execution**, **spoken-tag leakage**, or **multi-take directional contrast**.

---

## 3. Proposed `mizan-templates/packs/audio/` Additions

We propose adding three new templates to `packs/audio/templates/` in `mizan-templates`:

### 3.1 `audio/vocal-tag-adherence-boul` (`boul` — Binary Checklist)
- **Inputs (`modalities`):**
  - `recording` (`audio`): Synthesized `.wav` file.
  - `tagged_script` (`text`): Script containing inline tags (e.g., `<sigh> Fine, I'll do it <short pause> TOMORROW.`).
  - `expected_tags` (`text`): Comma-separated list of inline tags expected in the performance.
- **Checklist Criteria (`boul`):**
  1. `tag_performed`: Each inline paralinguistic tag (`<sigh>`, `<laugh>`, `<gasp>`, `<whispers>`, `<short pause>`, `<long pause>`) is audibly performed as a non-verbal vocalization, pause, or delivery shift at the corresponding point in the audio.
  2. `no_literal_tag_read_aloud`: The speaker does **not** pronounce the tag name itself as a spoken word (e.g., does not say the word "sigh", "laugh", "whispers", or "short pause").
  3. `emphasis_realized`: Words marked in `ALL-CAPS` receive audible prosodic stress (pitch, volume, or duration emphasis) relative to surrounding words.
  4. `transcript_complete`: All non-tag script words are spoken intelligibly without truncation or hallucinated words.

### 3.2 `audio/style-direction-adherence-rubric` (`rubric` — 1 to 5 Scale)
- **Inputs (`modalities`):**
  - `recording` (`audio`): Synthesized `.wav` file.
  - `style_directive` (`text`): Target acting direction (`persona`, `subtext`, `speechMetadata.style`, e.g., `"panicked, out of breath, speaking rapidly"`).
  - `plain_script` (`text`): Plain transcript without inline tags.
- **Rubric Anchors (1–5):**
  - **5 (Masterful Direction Match):** Tone, pacing, breathiness, and emotional subtext unmistakably embody the `style_directive` from the first syllable to the end.
  - **3 (Partial / Neutral Delivery):** Transcript is read clearly, but the emotional direction is muted, generic, or inconsistent with `style_directive`.
  - **1 (Contradictory or Flat):** Delivery contradicts the requested style (e.g., upbeat and slow when `"panicked, speaking rapidly"` was requested) or reads stage directions aloud.

### 3.3 `audio/threeup-take-contrast-custom` (`custom_schema`)
- **Inputs (`modalities`):**
  - `take_1` (`audio`): Safe / Commercial Baseline `.wav`.
  - `take_2` (`audio`): Pushed / Editorial `.wav`.
  - `take_3` (`audio`): Wildcard / Theatrical `.wav`.
  - `script` (`text`): Base script.
- **Output Schema:**
  - `takes_audibly_distinct` (`boolean`): Whether all three takes sound like genuinely different acting choices rather than near-identical reads.
  - `energy_progression_valid` (`boolean`): Whether vocal intensity and stylistic risk escalate from `take_1 -> take_2 -> take_3`.
  - `Literal_tag_leakage_detected` (`boolean`): Whether any take spoke a tag name aloud.
  - `summary` (`string`): Brief studio director critique comparing the three takes.

---

## 4. Proposed Harness in `take3bounce` (`backend/tools/eval_speech_tags/`)

1. **Golden Fixture Set (`backend/tools/eval_speech_tags/fixtures.json`):**
   - 10 representative short scripts covering every supported inline vocal tag (`<sigh>`, `<laugh>`, `<chuckle>`, `<gasp>`, `<pant>`, `<whispers>`, `<short pause>`, `<long pause>`, `ALL-CAPS` stress) and contrasting `speechMetadata.style` directives.
2. **Fixture Synthesizer (`backend/tools/eval_speech_tags/main.go`):**
   - Synthesizes `.wav` files for each fixture against the configured `GEMINI_TTS_MODEL` (`gemini-3.1-flash-tts-preview` or `gemini-3.8-flash-tts`) and emits a `mizan`-compatible JSONL dataset (`dataset.jsonl`).
3. **Evaluation Runner Target (`Makefile`):**
   - `make eval-tts`: Runs `mizan eval` against `dataset.jsonl` using `audio/vocal-tag-adherence-boul` and `audio/style-direction-adherence-rubric`, failing if `no_literal_tag_read_aloud` is `false` on any fixture.
